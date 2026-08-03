package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/infra/config"
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
)

func TestHandler_loginMeLogout_usesOpaqueSessionCookie(t *testing.T) {
	service, client, csrf := newHandlerTestService(t, "user")
	handler := NewHandler(service)
	login := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"correct horse battery staple"}`, csrf, csrf.Value)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}
	if strings.Contains(login.Body.String(), "session") || strings.Contains(login.Body.String(), "token") {
		t.Fatalf("login response exposed session material: %s", login.Body.String())
	}
	session := responseCookie(login, service.SessionCookieName())
	if session == nil {
		t.Fatal("login did not set a session cookie")
	}
	if !session.HttpOnly || !session.Secure || session.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session cookie flags = %+v", session)
	}
	stored, err := client.Session.Query().Only(context.Background())
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if stored.SessionToken == session.Value {
		t.Fatal("session token was stored in plaintext")
	}
	me := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(session)
	handler.ServeHTTP(me, request)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"user@example.com"`) {
		t.Fatalf("me response = %d %s", me.Code, me.Body.String())
	}
	logout := doRequest(t, handler, http.MethodPost, "/api/auth/logout", session, csrf)
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status = %d, body = %s", logout.Code, logout.Body.String())
	}
	cleared := responseCookie(logout, service.SessionCookieName())
	if cleared == nil || cleared.MaxAge >= 0 {
		t.Fatal("logout did not expire the session cookie")
	}
	stale := httptest.NewRecorder()
	staleReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	staleReq.AddCookie(session)
	handler.ServeHTTP(stale, staleReq)
	if stale.Code != http.StatusUnauthorized {
		t.Fatalf("stale session status = %d", stale.Code)
	}
}

func TestHandler_requiresCSRFForStateChanges(t *testing.T) {
	service, _, _ := newHandlerTestService(t, "user")
	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/logout", strings.NewReader(`{}`)))
	if response.Code != http.StatusForbidden {
		t.Fatalf("missing csrf status = %d", response.Code)
	}
}

func TestHandler_rateLimitsFailedLogins(t *testing.T) {
	service, _, csrf := newHandlerTestService(t, "user", serviceauth.WithLoginRateLimit(2, time.Minute))
	handler := NewHandler(service)
	for attempt := 0; attempt < 2; attempt++ {
		response := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"wrong"}`, csrf, csrf.Value)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("failed attempt %d status = %d", attempt, response.Code)
		}
	}
	limited := doJSONRequest(t, handler, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"wrong"}`, csrf, csrf.Value)
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("rate-limited status = %d", limited.Code)
	}
}

func newHandlerTestService(t *testing.T, role string, options ...serviceauth.Option) (*serviceauth.Service, *ent.Client, *http.Cookie) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/auth.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	hash, err := serviceauth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.User.Create().SetID("user-1").SetEmail("user@example.com").SetUsername("user").SetPassword(hash).SetRole(role).SetCredits(100).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	options = append([]serviceauth.Option{serviceauth.WithSecureCookies(true)}, options...)
	service, err := serviceauth.NewService(client, config.Auth{SessionSecret: "test-session-secret"}, options...)
	if err != nil {
		t.Fatal(err)
	}
	csrfResponse := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(csrfResponse, httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil))
	csrf := responseCookie(csrfResponse, service.CSRFTokenCookieName())
	if csrf == nil {
		t.Fatal("csrf endpoint did not set a cookie")
	}
	return service, client, csrf
}

func doJSONRequest(t *testing.T, handler http.Handler, method, path, body string, csrfCookie *http.Cookie, csrfToken string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(middleware.CSRFHeaderName, csrfToken)
	request.AddCookie(csrfCookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
func doRequest(t *testing.T, handler http.Handler, method, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	request.Header.Set(middleware.CSRFHeaderName, cookies[len(cookies)-1].Value)
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
