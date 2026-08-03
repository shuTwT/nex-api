package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
)

func TestHandler_rejectsUnknownJSONFields(t *testing.T) {
	mux := newCatalogMux(t)
	request := adminRequest(http.MethodPost, "/api/categories", `{"name":"tools","unexpected":true}`)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
func TestCatalogRoutesRequireAdmin(t *testing.T) {
	mux := newCatalogMux(t)
	for _, path := range []string{"/api/apis", "/api/categories", "/api/mcp-services"} {
		t.Run(path, func(t *testing.T) {
			for _, test := range []struct {
				name, role string
				status     int
			}{
				{"unauthenticated", "", http.StatusUnauthorized}, {"user", "user", http.StatusForbidden}, {"admin", "admin", http.StatusOK},
			} {
				t.Run(test.name, func(t *testing.T) {
					request := httptest.NewRequest(http.MethodGet, path, nil)
					if test.role != "" {
						request = request.WithContext(serviceauth.WithAuthContext(request.Context(), serviceauth.AuthContext{User: serviceauth.User{ID: test.role + "-1", Role: test.role}}))
					}
					response := httptest.NewRecorder()
					mux.ServeHTTP(response, request)
					if response.Code != test.status {
						t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
					}
				})
			}
		})
	}
}
func TestCatalogRoutes_acceptAllFiltersAndEmptyStats(t *testing.T) {
	mux := newCatalogMux(t)
	for _, path := range []string{"/api/apis/stats", "/api/apis?category=all&search=&status=all&page=1&limit=10", "/api/mcp-services?type=all&search=&status=all&page=1&limit=10"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, adminRequest(http.MethodGet, path, ""))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}
func TestCatalogResponseShape_includesEnvelope(t *testing.T) {
	mux := newCatalogMux(t)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, adminRequest(http.MethodGet, "/api/categories", ""))
	var body struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || string(body.Data) != "[]" {
		t.Fatalf("body = %s", body.Data)
	}
}

func newCatalogMux(t *testing.T) *http.ServeMux {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/catalog.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	apis, err := servicecatalog.NewAPIService(client)
	if err != nil {
		t.Fatal(err)
	}
	categories, err := servicecatalog.NewCategoryService(client)
	if err != nil {
		t.Fatal(err)
	}
	mcp, err := servicecatalog.NewMCPService(client)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(apis, categories, mcp)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}
	return mux
}
func adminRequest(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	return request.WithContext(serviceauth.WithAuthContext(context.Background(), serviceauth.AuthContext{User: serviceauth.User{ID: "admin-1", Role: "admin"}}))
}
