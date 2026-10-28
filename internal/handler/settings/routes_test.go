package settings

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	serviceauth "github.com/shuTwT/nex-api/internal/service/auth"
	servicesettings "github.com/shuTwT/nex-api/internal/service/settings"
)

var adminRoutes = []struct{ method, path string }{
	{http.MethodGet, "/api/system-settings"},
	{http.MethodPut, "/api/system-settings"},
}

func newTestMux(t *testing.T) *chi.Mux {
	t.Helper()
	mux := chi.NewRouter()
	// Requests below are rejected or fail validation before any service call,
	// so a zero-value service is enough.
	if err := RegisterRoutes(mux, &Handler{service: &servicesettings.Service{}}); err != nil {
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

// An admin must get past the middleware and reach the handler. A malformed body
// makes the handler answer 400 before it touches the service.
func TestRegisterRoutes_adminReachesHandler(t *testing.T) {
	mux := newTestMux(t)
	recorder := httptest.NewRecorder()
	request := withUser(httptest.NewRequest(http.MethodPut, "/api/system-settings", strings.NewReader("not json")), "admin-1", "admin")
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 from the handler; body=%s", recorder.Code, recorder.Body.String())
	}
}
