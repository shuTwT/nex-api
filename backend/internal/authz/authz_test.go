package authz

import (
	"bytes"
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
	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

const authzSchema = `CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT, "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT, "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL DEFAULT 'user', "credits" INTEGER NOT NULL DEFAULT 1000, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL); CREATE TABLE "ApiToken" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT NOT NULL, "name" TEXT NOT NULL, "token" TEXT NOT NULL UNIQUE, "permissions" TEXT NOT NULL, "lastUsedAt" DATETIME, "expiresAt" DATETIME, "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);`

func TestGenerateToken_hasExactFormat(t *testing.T) {
	// Given
	random := bytes.NewReader(bytes.Repeat([]byte{0xab}, TokenEntropySize))

	// When
	token, err := generateToken(random)

	// Then
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token != "sk_"+strings.Repeat("ab", TokenEntropySize) || !IsGeneratedToken(token) {
		t.Fatalf("token format = %q", token)
	}
}

func TestParseBearerToken_rejectsNonBearerHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
		valid  bool
	}{
		{name: "case insensitive bearer", header: "bEaReR sk_value", valid: true},
		{name: "missing header", header: "", valid: false},
		{name: "wrong scheme", header: "Basic sk_value", valid: false},
		{name: "extra values", header: "Bearer sk_value extra", valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			token, err := ParseBearerToken(test.header)

			// Then
			if test.valid && (err != nil || token != "sk_value") {
				t.Fatalf("parsed bearer = %q, error = %v", token, err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidBearer) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPermissions_mapMethodsToScopes(t *testing.T) {
	tests := []struct {
		permissions string
		method      string
		allowed     bool
	}{
		{permissions: "read", method: http.MethodGet, allowed: true},
		{permissions: "read", method: http.MethodPost, allowed: false},
		{permissions: "read,write", method: http.MethodPatch, allowed: true},
		{permissions: "read,write", method: http.MethodDelete, allowed: false},
		{permissions: "read,write,delete", method: http.MethodDelete, allowed: true},
	}
	for _, test := range tests {
		permissions, err := ParsePermissions(test.permissions)
		if err != nil {
			t.Fatalf("parse permissions: %v", err)
		}
		if permissions.AllowsMethod(test.method) != test.allowed {
			t.Errorf("%q %s allowed = %v", test.permissions, test.method, !test.allowed)
		}
	}
	for _, raw := range []string{"write", "read,delete"} {
		if _, err := ParsePermissions(raw); !errors.Is(err, ErrInvalidPermissions) {
			t.Errorf("permissions %q error = %v", raw, err)
		}
	}
}

func TestTokenService_authenticatesAndAtomicallyTouchesActiveToken(t *testing.T) {
	// Given
	service, client, now := newTokenService(t)
	seedToken(t, client, "token-1", "user-1", "sk_active", "read", true, time.Time{})

	// When
	principal, err := service.AuthenticateBearer(context.Background(), "Bearer sk_active")

	// Then
	if err != nil {
		t.Fatalf("authenticate token: %v", err)
	}
	if principal.UserID != "user-1" || principal.Source != APITokenCredential {
		t.Fatalf("principal = %+v", principal)
	}
	stored, err := client.ApiToken.Get(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("query token: %v", err)
	}
	if !stored.LastUsedAt.Equal(now) {
		t.Fatalf("lastUsedAt = %v, want %v", stored.LastUsedAt, now)
	}
}

func TestTokenService_rejectsInactiveExpiredAndInvalidPermissionTokens(t *testing.T) {
	// Given
	service, client, now := newTokenService(t)
	seedToken(t, client, "inactive", "user-1", "sk_inactive", "read", false, time.Time{})
	seedToken(t, client, "expired", "user-1", "sk_expired", "read", true, now)
	seedToken(t, client, "malformed", "user-1", "sk_malformed", "root", true, time.Time{})

	for _, token := range []string{"sk_inactive", "sk_expired", "sk_malformed", "sk_missing"} {
		// When
		_, err := service.Authenticate(context.Background(), token)

		// Then
		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("token %q error = %v", token, err)
		}
	}
}

func TestAPITokenMiddleware_distinguishesAuthenticationAndPermissionFailures(t *testing.T) {
	// Given
	service, client, _ := newTokenService(t)
	seedToken(t, client, "read-token", "user-1", "sk_read", "read", true, time.Time{})
	handler := service.RequireAPIToken(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name   string
		method string
		header string
		status int
	}{
		{name: "missing bearer", method: http.MethodGet, status: http.StatusUnauthorized},
		{name: "read request", method: http.MethodGet, header: "Bearer sk_read", status: http.StatusNoContent},
		{name: "write denied by read token", method: http.MethodPost, header: "Bearer sk_read", status: http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/api/v1/demo", nil)
			if test.header != "" {
				request.Header.Set("Authorization", test.header)
			}
			response := httptest.NewRecorder()

			// When
			handler.ServeHTTP(response, request)

			// Then
			if response.Code != test.status {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestBrowserPolicies_doNotTrustAPITokenContext(t *testing.T) {
	// Given
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/api/tokens", nil)
	request = request.WithContext(WithPrincipal(request.Context(), Principal{UserID: "user-1", Source: APITokenCredential}))
	response := httptest.NewRecorder()

	// When
	RequireUser(next).ServeHTTP(response, request)

	// Then
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestAdminMiddleware_returnsForbiddenForUser(t *testing.T) {
	// Given
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	request = request.WithContext(auth.WithAuthContext(request.Context(), auth.AuthContext{
		User: auth.User{ID: "user-1", Role: "user"},
	}))
	response := httptest.NewRecorder()

	// When
	RequireAdmin(next).ServeHTTP(response, request)

	// Then
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestOwnershipMiddleware_forbidsCrossUserAndAllowsAdmin(t *testing.T) {
	owner := func(*http.Request) (string, error) { return "user-2", nil }
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	tests := []struct {
		name   string
		role   string
		status int
	}{
		{name: "different user", role: "user", status: http.StatusForbidden},
		{name: "admin", role: "admin", status: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/tokens/token-2", nil)
			request = request.WithContext(auth.WithAuthContext(request.Context(), auth.AuthContext{
				User: auth.User{ID: "user-1", Role: test.role},
			}))
			response := httptest.NewRecorder()

			// When
			RequireOwnership(owner)(next).ServeHTTP(response, request)

			// Then
			if response.Code != test.status {
				t.Fatalf("status = %d", response.Code)
			}
		})
	}
}

func newTokenService(t *testing.T) (*TokenService, *ent.Client, time.Time) {
	databasePath := filepath.Join(t.TempDir(), "authz.db")
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	_, err = database.ExecContext(context.Background(), authzSchema)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close schema database: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, databasePath)
	if err != nil {
		t.Fatalf("open ent client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	for _, account := range []struct{ id, role string }{{"user-1", "user"}, {"user-2", "admin"}} {
		_, err = client.User.Create().SetID(account.id).SetEmail(account.id + "@example.com").SetUsername(account.id).SetPassword("redacted").SetRole(account.role).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background())
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	store, err := NewEntTokenStore(client)
	if err != nil {
		t.Fatalf("create token store: %v", err)
	}
	service, err := NewTokenService(store, WithClock(clockFunc(func() time.Time { return now })))
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}
	return service, client, now
}

func seedToken(t *testing.T, client *ent.Client, id, userID, token, permissions string, active bool, expiresAt time.Time) {
	builder := client.ApiToken.Create().SetID(id).SetUserID(userID).SetName(id).SetToken(token).SetPermissions(permissions).SetIsActive(active).SetCreatedAt(time.Now()).SetUpdatedAt(time.Now())
	if !expiresAt.IsZero() {
		builder.SetExpiresAt(expiresAt)
	}
	if _, err := builder.Save(context.Background()); err != nil {
		t.Fatalf("create token: %v", err)
	}
}
