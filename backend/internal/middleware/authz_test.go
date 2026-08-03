package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
)

type tokenStoreStub struct{ token serviceauthz.StoredToken }

func (s tokenStoreStub) LookupToken(_ context.Context, token string) (serviceauthz.StoredToken, error) {
	if token != s.token.Token {
		return serviceauthz.StoredToken{}, serviceauthz.ErrTokenNotFound
	}
	return s.token, nil
}
func (tokenStoreStub) TouchLastUsedAt(context.Context, string, time.Time) error { return nil }

func TestAPITokenMiddleware_distinguishesAuthenticationAndPermissionFailures(t *testing.T) {
	service, err := serviceauthz.NewTokenService(tokenStoreStub{token: serviceauthz.StoredToken{ID: "read-token", UserID: "user-1", Token: "sk_read", Permissions: "read", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	handler := RequireAPIToken(service, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, test := range []struct {
		name, method, header string
		status               int
	}{{"missing bearer", http.MethodGet, "", http.StatusUnauthorized}, {"read request", http.MethodGet, "Bearer sk_read", http.StatusNoContent}, {"write denied by read token", http.MethodPost, "Bearer sk_read", http.StatusForbidden}} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/api/v1/demo", nil)
			if test.header != "" {
				request.Header.Set("Authorization", test.header)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestBrowserPolicies_doNotTrustAPITokenContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/api/tokens", nil)
	request = request.WithContext(serviceauthz.WithPrincipal(request.Context(), serviceauthz.Principal{UserID: "user-1", Source: serviceauthz.APITokenCredential}))
	response := httptest.NewRecorder()
	RequireUser(next).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}
func TestAdminMiddleware_returnsForbiddenForUser(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	request = request.WithContext(serviceauth.WithAuthContext(request.Context(), serviceauth.AuthContext{User: serviceauth.User{ID: "user-1", Role: "user"}}))
	response := httptest.NewRecorder()
	RequireAdmin(next).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestRequireAdmin_populatesAuthorizationContext(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request = request.WithContext(serviceauth.WithAuthContext(request.Context(), serviceauth.AuthContext{User: serviceauth.User{ID: "admin-1", Role: "admin"}}))
	response := httptest.NewRecorder()
	RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := serviceauthz.PrincipalFromContext(r.Context())
		if !ok || principal.UserID != "admin-1" || principal.Role != "admin" || principal.Source != serviceauthz.BrowserSessionCredential {
			t.Errorf("authorization principal = %+v, ok = %v", principal, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("admin status = %d", response.Code)
	}
}
func TestOwnershipMiddleware_forbidsCrossUserAndAllowsAdmin(t *testing.T) {
	owner := func(*http.Request) (string, error) { return "user-2", nil }
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	for _, test := range []struct {
		name, role string
		status     int
	}{{"different user", "user", http.StatusForbidden}, {"admin", "admin", http.StatusNoContent}} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/tokens/token-2", nil)
			request = request.WithContext(serviceauthz.WithPrincipal(request.Context(), serviceauthz.Principal{UserID: "user-1", Role: test.role, Source: serviceauthz.BrowserSessionCredential}))
			response := httptest.NewRecorder()
			RequireOwnership(owner)(next).ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d", response.Code)
			}
		})
	}
}

var _ = errors.Is
