package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

const legacyPasswordHash = "00112233445566778899aabbccddeeff:5699cfee2c5c280e66678242092f368ce88ff05305af2c75a9e629d473deb2165b3797e0e31ec3cda30414573befb697f928384c38b187e8c176107e5be20f01"
const authSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL DEFAULT '', "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT NOT NULL DEFAULT '', "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL, "credits" INTEGER NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL); CREATE TABLE "Session" ("id" TEXT PRIMARY KEY NOT NULL, "sessionToken" TEXT NOT NULL UNIQUE, "userId" TEXT NOT NULL, "expires" DATETIME NOT NULL, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestVerifyPassword_acceptsLegacyScryptEncoding(t *testing.T) {
	// Given
	hashed := legacyPasswordHash

	// When
	matched, err := VerifyPassword("correct horse battery staple", hashed)

	// Then
	if err != nil {
		t.Fatalf("VerifyPassword returned an error: %v", err)
	}
	if !matched {
		t.Fatal("VerifyPassword rejected the legacy hash")
	}
}

func TestVerifyPassword_rejectsWrongAndMalformedValues(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hashed   string
	}{
		{name: "wrong password", password: "wrong", hashed: legacyPasswordHash},
		{name: "missing separator", password: "correct horse battery staple", hashed: "not-a-hash"},
		{name: "wrong key length", password: "correct horse battery staple", hashed: "00112233445566778899aabbccddeeff:00"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			matched, err := VerifyPassword(test.password, test.hashed)

			// Then
			if err != nil {
				t.Fatalf("VerifyPassword returned an error: %v", err)
			}
			if matched {
				t.Fatal("VerifyPassword accepted an invalid value")
			}
		})
	}
}

func TestHandler_loginMeLogout_usesOpaqueSessionCookie(t *testing.T) {
	// Given
	service, client, csrfCookie := newTestService(t, "login-flow", "user")
	handler := NewHandler(service)

	// When: login
	loginResponse := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"correct horse battery staple"}`, csrfCookie, csrfCookie.Value)

	// Then
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResponse.Code, loginResponse.Body.String())
	}
	if strings.Contains(loginResponse.Body.String(), "session") || strings.Contains(loginResponse.Body.String(), "token") {
		t.Fatalf("login response exposed session material: %s", loginResponse.Body.String())
	}
	sessionCookie := responseCookie(loginResponse, service.SessionCookieName())
	if sessionCookie == nil {
		t.Fatal("login did not set a session cookie")
	}
	if !sessionCookie.HttpOnly || !sessionCookie.Secure || sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session cookie flags = %+v", sessionCookie)
	}
	storedSession, err := client.Session.Query().Only(context.Background())
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if storedSession.SessionToken == sessionCookie.Value {
		t.Fatal("session token was stored in plaintext")
	}

	// When: current-user lookup
	meRequest := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meRequest.AddCookie(sessionCookie)
	meResponse := httptest.NewRecorder()
	handler.ServeHTTP(meResponse, meRequest)

	// Then
	if meResponse.Code != http.StatusOK || !strings.Contains(meResponse.Body.String(), `"user@example.com"`) {
		t.Fatalf("me response = %d %s", meResponse.Code, meResponse.Body.String())
	}

	// When: logout
	logoutResponse := doRequest(t, handler, http.MethodPost, "/api/auth/logout", sessionCookie, csrfCookie)

	// Then
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("logout status = %d, body = %s", logoutResponse.Code, logoutResponse.Body.String())
	}
	clearedSession := responseCookie(logoutResponse, service.SessionCookieName())
	if clearedSession == nil || clearedSession.MaxAge >= 0 {
		t.Fatal("logout did not expire the session cookie")
	}
	meAfterLogout := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meAfterLogout.AddCookie(sessionCookie)
	meAfterLogoutResponse := httptest.NewRecorder()
	handler.ServeHTTP(meAfterLogoutResponse, meAfterLogout)
	if meAfterLogoutResponse.Code != http.StatusUnauthorized {
		t.Fatalf("stale session status = %d", meAfterLogoutResponse.Code)
	}
}

func TestHandler_requiresCSRFForStateChanges(t *testing.T) {
	// Given
	service, _, _ := newTestService(t, "csrf", "user")
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	// When
	NewHandler(service).ServeHTTP(response, request)

	// Then
	if response.Code != http.StatusForbidden {
		t.Fatalf("missing csrf status = %d", response.Code)
	}
}

func TestHandler_rateLimitsFailedLogins(t *testing.T) {
	// Given
	service, _, csrfCookie := newTestService(t, "rate-limit", "user", WithLoginRateLimit(2, time.Minute))
	handler := NewHandler(service)
	for attempt := 0; attempt < 2; attempt++ {
		response := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"wrong"}`, csrfCookie, csrfCookie.Value)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("failed attempt %d status = %d", attempt, response.Code)
		}
	}

	// When
	limited := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"wrong"}`, csrfCookie, csrfCookie.Value)

	// Then
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("rate-limited status = %d", limited.Code)
	}
}

func TestService_rotationInvalidatesPreviousSession(t *testing.T) {
	// Given
	service, _, _ := newTestService(t, "rotation", "user")
	first, err := service.Login(context.Background(), "user@example.com", "correct horse battery staple", "")
	if err != nil {
		t.Fatalf("initial login: %v", err)
	}

	// When
	rotated, err := service.RotateSession(context.Background(), first)

	// Then
	if err != nil {
		t.Fatalf("rotate session: %v", err)
	}
	if first.token == rotated.token {
		t.Fatal("session rotation reused the old token")
	}
	if _, err := service.Authenticate(context.Background(), first.token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("old session error = %v", err)
	}
	if _, err := service.Authenticate(context.Background(), rotated.token); err != nil {
		t.Fatalf("rotated session was not accepted: %v", err)
	}
}

func TestRequireAdmin_populatesAuthorizationContext(t *testing.T) {
	// Given
	service, _, _ := newTestService(t, "admin-context", "admin")
	authContext, err := service.Login(context.Background(), "user@example.com", "correct horse battery staple", "")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: service.SessionCookieName(), Value: authContext.token})
	response := httptest.NewRecorder()

	// When
	service.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user.Role != "admin" {
			t.Errorf("authorization context = %+v, ok = %v", user, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)

	// Then
	if response.Code != http.StatusNoContent {
		t.Fatalf("admin status = %d", response.Code)
	}
}

func newTestService(t *testing.T, name, role string, options ...Option) (*Service, *ent.Client, *http.Cookie) {
	databasePath := filepath.Join(t.TempDir(), name+".db")
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	_, err = database.ExecContext(context.Background(), authSchema)
	if err != nil {
		t.Fatalf("create auth tables: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, databasePath)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	_, err = client.User.Create().
		SetID("user-1").
		SetEmail("user@example.com").
		SetUsername("user").
		SetPassword(legacyPasswordHash).
		SetRole(role).
		SetCredits(100).
		SetCreatedAt(time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)).
		SetUpdatedAt(time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)).
		Save(context.Background())
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	options = append([]Option{WithClock(clockFunc(func() time.Time { return time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) }))}, options...)
	service, err := NewService(client, config.Auth{SessionSecret: "test-session-secret"}, options...)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	csrfRequest := httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil)
	csrfResponse := httptest.NewRecorder()
	service.Handler().ServeHTTP(csrfResponse, csrfRequest)
	csrfCookie := responseCookie(csrfResponse, service.CSRFTokenCookieName())
	if csrfCookie == nil {
		t.Fatal("csrf endpoint did not set a cookie")
	}
	return service, client, csrfCookie
}

func doJSONRequest(t *testing.T, handler http.Handler, method, path, body string, csrfCookie *http.Cookie, csrfToken string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(CSRFHeaderName, csrfToken)
	request.AddCookie(csrfCookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func doRequest(t *testing.T, handler http.Handler, method, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	request.Header.Set(CSRFHeaderName, cookies[len(cookies)-1].Value)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func responseCookie(response *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
