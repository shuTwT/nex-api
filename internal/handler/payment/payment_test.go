package payment

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/enttest"
	"github.com/shuTwT/nex-api/ent/payment"
	pay "github.com/shuTwT/nex-api/internal/infra/pay"
	servicepayment "github.com/shuTwT/nex-api/internal/service/payment"
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
	for _, path := range []string{"/api/recharge", "/api/payment/orders"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("POST %s status = %d", path, response.Code)
			}
		})
	}
}
func TestRegisterRoutesRejectsOrderCreationOnMethodsList(t *testing.T) {
	mux := newPaymentMux(t)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/payment/methods", strings.NewReader(`{}`)))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/payment/methods status = %d, want 405", response.Code)
	}
}
func newPaymentMux(t *testing.T) chi.Router {
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
	mux := chi.NewRouter()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}
	return mux
}

// epayTestSign recomputes the V1 signature from the documented algorithm
// (sorted non-empty k=v pairs joined by &, then the merchant key, MD5 lowercase)
// instead of calling the provider helper, so a broken canonicalization fails.
func epayTestSign(values url.Values, key string) string {
	names := make([]string, 0, len(values))
	for name, field := range values {
		if name == "sign" || name == "sign_type" || len(field) == 0 || field[0] == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+"="+values.Get(name))
	}
	digest := md5.Sum([]byte(strings.Join(parts, "&") + key))
	return hex.EncodeToString(digest[:])
}

func newEpayCallbackFixture(t *testing.T, outTradeNo string) (chi.Router, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/payment.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	now := time.Now().UTC()
	if _, err := client.User.Create().SetID("u-1").SetName("User").SetEmail("u-1@example.com").SetUsername("u-1").SetPassword("redacted").SetRole("user").SetCredits(1000).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	metadata := `{"type":"recharge","credits":10}`
	if _, err := client.Payment.Create().SetUserId("u-1").SetOutTradeNo(outTradeNo).SetMethod("alipay").SetAmount(10).SetCurrency("CNY").SetStatus("pending").SetCreatedAt(now).SetUpdatedAt(now).SetExpiredAt(now.Add(2 * time.Hour)).SetMetadata(metadata).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	service, err := servicepayment.NewServiceWithProviders(client, []pay.Provider{pay.NewEpayProvider(pay.EpayConfiguration{Enabled: true, APIURL: "https://epay.example.com", PID: "1001", Key: "merchant-key", Version: pay.EpayVersionV1}, pay.PaymentMethodAlipay, nil)})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	mux := chi.NewRouter()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}
	return mux, client
}

func epayNotifyForm(outTradeNo, sign string) url.Values {
	return url.Values{
		"pid":          {"1001"},
		"trade_no":     {"2026080622555342651"},
		"out_trade_no": {outTradeNo},
		"type":         {"alipay"},
		"name":         {"积分充值"},
		"money":        {"10.00"},
		"trade_status": {"TRADE_SUCCESS"},
		"sign_type":    {"MD5"},
		"sign":         {sign},
	}
}

func TestEpayCallbackViaPostMarksOrderPaid(t *testing.T) {
	mux, client := newEpayCallbackFixture(t, "PAY-POST1")
	form := epayNotifyForm("PAY-POST1", "")
	form.Set("sign", epayTestSign(form, "merchant-key"))
	request := httptest.NewRequest(http.MethodPost, "/api/payment/callback/alipay", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "success" {
		t.Fatalf("POST callback status=%d body=%q", response.Code, response.Body.String())
	}
	record, err := client.Payment.Query().Where(payment.OutTradeNo("PAY-POST1")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != "paid" || record.TransactionId != "2026080622555342651" {
		t.Fatalf("record = %+v", record)
	}
	user, err := client.User.Get(context.Background(), "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 1010 {
		t.Fatalf("credits = %d", user.Credits)
	}
}

func TestEpayCallbackViaGetMarksOrderPaid(t *testing.T) {
	mux, client := newEpayCallbackFixture(t, "PAY-GET1")
	form := epayNotifyForm("PAY-GET1", "")
	form.Set("sign", epayTestSign(form, "merchant-key"))
	request := httptest.NewRequest(http.MethodGet, "/api/payment/callback/alipay?"+form.Encode(), nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "success" {
		t.Fatalf("GET callback status=%d body=%q", response.Code, response.Body.String())
	}
	record, err := client.Payment.Query().Where(payment.OutTradeNo("PAY-GET1")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != "paid" {
		t.Fatalf("record = %+v", record)
	}
}

func TestEpayCallbackRejectsBadSignature(t *testing.T) {
	mux, client := newEpayCallbackFixture(t, "PAY-BADSIGN")
	form := epayNotifyForm("PAY-BADSIGN", "deadbeefdeadbeefdeadbeefdeadbeef")
	request := httptest.NewRequest(http.MethodPost, "/api/payment/callback/alipay", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code == http.StatusOK {
		t.Fatalf("expected forged callback to be rejected, got %d", response.Code)
	}
	record, err := client.Payment.Query().Where(payment.OutTradeNo("PAY-BADSIGN")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != "pending" {
		t.Fatalf("forged callback must not change status, got %s", record.Status)
	}
}
