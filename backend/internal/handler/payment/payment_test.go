package payment

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	pay "github.com/shuTwT/nex-api/backend/internal/infra/pay"
	servicepayment "github.com/shuTwT/nex-api/backend/internal/service/payment"
)

func TestRegisterRoutesDoesNotExposeGenericPaymentCreation(t *testing.T) {
	mux := newPaymentMux(t)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/payment", strings.NewReader(`{"amount":0.01,"method":"mock","metadata":{"type":"recharge","credits":1000000}}`)))
	if response.Code != http.StatusNotFound && response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/payment status = %d", response.Code)
	}
}
func TestRegisterRoutesKeepsBusinessPaymentCreationEndpoints(t *testing.T) {
	mux := newPaymentMux(t)
	for _, path := range []string{"/api/recharge", "/api/payment/methods"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("POST %s status = %d", path, response.Code)
			}
		})
	}
}
func newPaymentMux(t *testing.T) *http.ServeMux {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/payment.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	service, err := servicepayment.NewServiceWithProviders(client, []pay.Provider{pay.NewMockProvider(pay.MockPaymentConfiguration{Enabled: true})})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}
	return mux
}
