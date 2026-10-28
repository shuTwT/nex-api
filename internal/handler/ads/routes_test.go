package ads

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	serviceads "github.com/shuTwT/nex-api/internal/service/ads"
	serviceauth "github.com/shuTwT/nex-api/internal/service/auth"
)

var adminRoutes = []struct{ method, path string }{
	{http.MethodGet, "/api/advertisements"},
	{http.MethodPost, "/api/advertisements"},
	{http.MethodGet, "/api/advertisements/stats"},
	{http.MethodGet, "/api/advertisements/ad-1"},
	{http.MethodPut, "/api/advertisements/ad-1"},
	{http.MethodDelete, "/api/advertisements/ad-1"},
	{http.MethodPut, "/api/advertisements/ad-1/toggle"},
}

func newTestMux(t *testing.T) *chi.Mux {
	t.Helper()
	mux := chi.NewRouter()
	// Requests below are rejected or fail validation before any service call,
	// so a zero-value service is enough.
	if err := RegisterRoutes(mux, &Handler{service: &serviceads.Service{}}); err != nil {
		t.Fatal(err)
	}
	return mux
}

func withUser(r *http.Request, id, role string) *http.Request {
	authContext := serviceauth.AuthContext{User: serviceauth.User{ID: id, Role: role}}
	return r.WithContext(serviceauth.WithAuthContext(r.Context(), authContext))
}

func TestRegisterRoutes_adminRoutesRejectAnonymous(t *testing.T) {
	mux := newTestMux(t)
	for _, route := range adminRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestRegisterRoutes_adminRoutesRejectNonAdmin(t *testing.T) {
	mux := newTestMux(t)
	for _, route := range adminRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := withUser(httptest.NewRequest(route.method, route.path, nil), "user-1", "user")
			mux.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

// An admin must get past the middleware and reach the handler. An invalid page
// makes the handler answer 400 before it touches the service.
func TestRegisterRoutes_adminReachesHandler(t *testing.T) {
	mux := newTestMux(t)
	recorder := httptest.NewRecorder()
	request := withUser(httptest.NewRequest(http.MethodGet, "/api/advertisements?page=abc", nil), "admin-1", "admin")
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 from the handler; body=%s", recorder.Code, recorder.Body.String())
	}
}
