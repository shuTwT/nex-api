package payment

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/payment"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/systemsetting"
)

func TestVerifyWeChatSignatureRejectsForgedCallback(t *testing.T) {
	// Given
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)})
	body := []byte(`{"resource":{"ciphertext":"not-used"}}`)
	timestamp := "1700000000"
	nonce := "nonce"
	headers := http.Header{
		"Wechatpay-Timestamp": []string{timestamp},
		"Wechatpay-Nonce":     []string{nonce},
		"Wechatpay-Signature": []string{"forged"},
	}

	// When
	err = VerifyWeChatSignature(body, headers, string(publicKeyPEM))

	// Then
	if err == nil {
		t.Fatal("expected forged signature to be rejected")
	}
}

func TestVerifyWeChatSignatureAcceptsPlatformCertificate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(42), Subject: pkix.Name{CommonName: "WeChat Pay Platform"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature,
	}
	certificate, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	body := []byte(`{"event_type":"TRANSACTION.SUCCESS"}`)
	timestamp := "1700000000"
	nonce := "nonce"
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	headers := http.Header{
		"Wechatpay-Timestamp": []string{timestamp},
		"Wechatpay-Nonce":     []string{nonce},
		"Wechatpay-Signature": []string{base64.StdEncoding.EncodeToString(signature)},
	}

	if err := VerifyWeChatSignature(body, headers, string(certificatePEM)); err != nil {
		t.Fatalf("verify platform certificate signature: %v", err)
	}
}

func TestWeChatCallbackSelectsPaymentPublicKeyBySerial(t *testing.T) {
	provider := NewWeChatProvider(weChatConfiguration{
		PublicKey: "platform-certificate", PaymentPublicKey: "payment-public-key", PublicKeyID: "PUB_KEY_ID_1",
	}, nil)

	key, err := provider.callbackVerificationKey(http.Header{"Wechatpay-Serial": []string{"PUB_KEY_ID_1"}})
	if err != nil {
		t.Fatal(err)
	}
	if key != "payment-public-key" {
		t.Fatalf("verification key = %q, want payment public key", key)
	}
	key, err = provider.callbackVerificationKey(http.Header{"Wechatpay-Serial": []string{"CERT_SERIAL"}})
	if err != nil {
		t.Fatal(err)
	}
	if key != "platform-certificate" {
		t.Fatalf("verification key = %q, want platform certificate", key)
	}
}

func TestWeChatConfigurationAllowsPaymentPublicKeyMode(t *testing.T) {
	configuration := weChatConfiguration{
		AppID: "app-id", MchID: "merchant-id", APIKey: "01234567890123456789012345678901",
		PrivateKey: "merchant-private-key", PaymentPublicKey: "payment-public-key", PublicKeyID: "PUB_KEY_ID_1",
	}
	if !wechatConfigured(configuration) {
		t.Fatal("payment public key mode should be considered configured without a platform certificate")
	}
}

func TestVerifyAlipaySignatureAcceptsSignedPayloadAndRejectsMutation(t *testing.T) {
	// Given
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)})
	values := url.Values{
		"out_trade_no": {"ALI-1"},
		"trade_no":     {"TX-1"},
		"total_amount": {"10.00"},
		"trade_status": {"TRADE_SUCCESS"},
	}
	signature, err := signAlipayValues(values, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	values.Set("sign", signature)

	// When
	err = VerifyAlipaySignature(values, string(publicKeyPEM))
	if err != nil {
		t.Fatalf("expected valid signature: %v", err)
	}
	values.Set("total_amount", "11.00")
	err = VerifyAlipaySignature(values, string(publicKeyPEM))

	// Then
	if err == nil {
		t.Fatal("expected mutated payload to be rejected")
	}
}

func TestComparePaymentAmountRequiresExactCentsAndCurrency(t *testing.T) {
	// Given / When / Then
	if err := ComparePaymentAmount(10, "CNY", 10.00, "CNY"); err != nil {
		t.Fatalf("expected matching amount: %v", err)
	}
	if err := ComparePaymentAmount(10, "CNY", 10.01, "CNY"); err == nil {
		t.Fatal("expected amount mismatch")
	}
	if err := ComparePaymentAmount(10, "CNY", 10, "USD"); err == nil {
		t.Fatal("expected currency mismatch")
	}
}

func TestPaymentStateTransitionDoesNotOverridePaidOrCancelled(t *testing.T) {
	// Given
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	// When / Then
	paid, err := transitionPayment(PaymentStatePending, PaymentStatePaid, now)
	if err != nil || paid.Status != PaymentStatePaid || paid.PaidAt != now {
		t.Fatalf("pending to paid transition failed: %#v %v", paid, err)
	}
	if _, err := transitionPayment(PaymentStatePaid, PaymentStateCancelled, now); err == nil {
		t.Fatal("expected paid payment to reject cancellation")
	}
	if _, err := transitionPayment(PaymentStateCancelled, PaymentStatePaid, now); err == nil {
		t.Fatal("expected cancelled payment to reject payment")
	}
}

func TestRegisterRoutesDoesNotExposeGenericPaymentCreation(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewServiceWithProviders(client, []Provider{NewMockProvider(mockPaymentConfiguration{Enabled: true})})
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

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/payment", strings.NewReader(`{"amount":0.01,"method":"mock","metadata":{"type":"recharge","credits":1000000}}`))
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("POST /api/payment status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestRegisterRoutesKeepsBusinessPaymentCreationEndpoints(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewServiceWithProviders(client, []Provider{NewMockProvider(mockPaymentConfiguration{Enabled: true})})
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

	for _, path := range []string{"/api/recharge", "/api/payment/methods"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
			mux.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("POST %s status = %d, want %d", path, response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestServiceCreatePaymentReadsCurrentNotifyURLFromDatabase(t *testing.T) {
	client := newPaymentTestClient(t)
	settings := map[string]string{
		"wechatPayAppId":      "app-id",
		"wechatPayMchId":      "merchant-id",
		"wechatPayApiKey":     "01234567890123456789012345678901",
		"wechatPayPrivateKey": "private-key",
		"wechatPayPublicKey":  "public-key",
		"wechatPayNotifyUrl":  "https://example.com/first-callback",
	}
	for key, value := range settings {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	provider := &recordingProvider{method: PaymentMethodWeChat}
	service.providerFactory = func(_ PaymentMethod, _ paymentConfiguration) Provider { return provider }

	if _, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: PaymentMethodWeChat}); err != nil {
		t.Fatal(err)
	}
	if provider.createRequest.NotifyURL != "https://example.com/first-callback" {
		t.Fatalf("notify URL = %q, want database value", provider.createRequest.NotifyURL)
	}

	setting, err := client.SystemSetting.Query().Where(systemsetting.Key("wechatPayNotifyUrl")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setting.Update().SetValue("https://example.com/updated-callback").Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 20, Currency: "CNY", Method: PaymentMethodWeChat}); err != nil {
		t.Fatal(err)
	}
	if provider.createRequest.NotifyURL != "https://example.com/updated-callback" {
		t.Fatalf("notify URL after settings update = %q, want current database value", provider.createRequest.NotifyURL)
	}
}

func TestPaymentConfigurationUsesAppURLFallbacks(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewService(client, "https://app.example.com/")
	if err != nil {
		t.Fatal(err)
	}

	configuration, err := service.loadConfiguration(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if configuration.WeChat.NotifyURL != "https://app.example.com/api/payment/callback/wechat" {
		t.Fatalf("wechat fallback URL = %q", configuration.WeChat.NotifyURL)
	}
	if configuration.Alipay.NotifyURL != "https://app.example.com/api/payment/callback/alipay" {
		t.Fatalf("alipay fallback URL = %q", configuration.Alipay.NotifyURL)
	}
	if configuration.Alipay.ReturnURL != "https://app.example.com/payment/result" {
		t.Fatalf("alipay return URL = %q", configuration.Alipay.ReturnURL)
	}
}

func TestMockPaymentAutoSuccessAndStatusQuery(t *testing.T) {
	client := newPaymentTestClient(t)
	for key, value := range map[string]string{
		"mockPaymentEnabled":     "true",
		"mockPaymentAutoSuccess": "true",
		"mockPaymentDelay":       "25",
	} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	var scheduledDelay time.Duration
	service.schedule = func(delay time.Duration, callback func()) {
		scheduledDelay = delay
		callback()
	}

	created, err := service.createPayment(context.Background(), createPaymentInput{
		UserID: "u-1", Amount: 10, Currency: "CNY", Method: PaymentMethodMock,
		Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.PayURL != "/payment/mock?outTradeNo="+url.QueryEscape(created.OutTradeNo) {
		t.Fatalf("mock pay URL = %q", created.PayURL)
	}
	if scheduledDelay != 25*time.Millisecond {
		t.Fatalf("mock delay = %s, want 25ms", scheduledDelay)
	}
	view, err := service.QueryPayment(context.Background(), created.OutTradeNo)
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != PaymentStatePaid {
		t.Fatalf("mock status = %s, want paid", view.Status)
	}
	user, err := client.User.Get(context.Background(), "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 1010 {
		t.Fatalf("credits = %d, want 1010", user.Credits)
	}
}

func TestMockPaymentFailureUsesFailedState(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewServiceWithProviders(client, []Provider{NewMockProvider(mockPaymentConfiguration{Enabled: true})})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: PaymentMethodMock})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"outTradeNo":"` + created.OutTradeNo + `","success":false}`)
	view, err := service.HandleCallback(context.Background(), PaymentMethodMock, ProviderCallbackRequest{Body: body})
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != PaymentStateFailed {
		t.Fatalf("mock failure status = %s, want failed", view.Status)
	}
}

func TestServiceHandleCallbackGrantsRechargeOnceForDuplicateCallbacks(t *testing.T) {
	// Given
	client := newPaymentTestClient(t)
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	service, err := NewServiceWithProviders(client, []Provider{NewMockProvider(mockPaymentConfiguration{Enabled: true})}, fixedClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: PaymentMethodMock, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")}})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"outTradeNo":"` + created.OutTradeNo + `","success":true,"amount":10,"currency":"CNY"}`)

	// When
	var wait sync.WaitGroup
	for range 4 {
		wait.Go(func() {
			if _, callbackErr := service.HandleCallback(context.Background(), PaymentMethodMock, ProviderCallbackRequest{Body: body}); callbackErr != nil {
				t.Errorf("duplicate callback: %v", callbackErr)
			}
		})
	}
	wait.Wait()

	// Then
	user, err := client.User.Get(context.Background(), "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 1010 {
		t.Fatalf("expected one credit grant, got %d", user.Credits)
	}
	paymentRecord, err := client.Payment.Query().Where(payment.OutTradeNo(created.OutTradeNo)).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if paymentRecord.Status != string(PaymentStatePaid) || paymentRecord.CallbackVersion != 1 {
		t.Fatalf("unexpected callback state: status=%s version=%d", paymentRecord.Status, paymentRecord.CallbackVersion)
	}
}

func TestExpirePendingPaymentsClosesProviderAndMarksOrderExpired(t *testing.T) {
	client := newPaymentTestClient(t)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	provider := &recordingProvider{method: PaymentMethodWeChat}
	service, err := NewServiceWithProviders(client, []Provider{provider}, fixedClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	order, err := client.Payment.Create().
		SetUserId("u-1").SetOutTradeNo("PAY-EXPIRED").SetMethod(string(PaymentMethodWeChat)).
		SetAmount(10).SetCurrency("CNY").SetStatus(string(PaymentStatePending)).
		SetExpiredAt(now.Add(-time.Minute)).SetCreatedAt(now.Add(-3 * time.Hour)).SetUpdatedAt(now.Add(-3 * time.Hour)).
		Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	count, err := service.ExpirePendingPayments(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(provider.cancelled) != 1 || provider.cancelled[0] != order.OutTradeNo {
		t.Fatalf("count=%d cancelled=%v", count, provider.cancelled)
	}
	updated, err := client.Payment.Get(context.Background(), order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != string(PaymentStateExpired) || updated.CallbackVersion != 1 {
		t.Fatalf("status=%s callbackVersion=%d", updated.Status, updated.CallbackVersion)
	}
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type recordingProvider struct {
	method        PaymentMethod
	createRequest ProviderCreateRequest
	cancelled     []string
}

func (p *recordingProvider) Method() PaymentMethod { return p.method }

func (p *recordingProvider) Create(_ context.Context, request ProviderCreateRequest) (ProviderCreateResult, error) {
	p.createRequest = request
	return ProviderCreateResult{PayURL: "https://pay.example.com"}, nil
}

func (p *recordingProvider) VerifyCallback(_ context.Context, _ ProviderCallbackRequest) (ProviderCallback, error) {
	return ProviderCallback{}, nil
}

func (p *recordingProvider) Query(_ context.Context, _ string) (ProviderStatus, error) {
	return ProviderStatus{Status: PaymentStatePending}, nil
}

func (p *recordingProvider) Cancel(_ context.Context, outTradeNo string) error {
	p.cancelled = append(p.cancelled, outTradeNo)
	return nil
}

func newPaymentTestClient(t *testing.T) *ent.Client {
	t.Helper()
	path := filepath.Join(t.TempDir(), "payment.db")
	database, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(paymentTestSchema); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	client, err := ent.Open(dialect.SQLite, path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	if _, err := client.User.Create().SetID("u-1").SetName("User").SetEmail("u-1@example.com").SetUsername("u-1").SetPassword("redacted").SetRole("user").SetCredits(1000).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	return client
}

const paymentTestSchema = `
CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT, "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT, "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL DEFAULT 'user', "credits" INTEGER NOT NULL DEFAULT 1000, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "Payment" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT NOT NULL, "outTradeNo" TEXT NOT NULL UNIQUE, "transactionId" TEXT, "method" TEXT NOT NULL, "amount" REAL NOT NULL, "currency" TEXT NOT NULL DEFAULT 'CNY', "status" TEXT NOT NULL DEFAULT 'pending', "qrcodeUrl" TEXT, "payUrl" TEXT, "notifyUrl" TEXT, "paidAt" DATETIME, "expiredAt" DATETIME, "cancelledAt" DATETIME, "metadata" TEXT, "callbackVersion" INTEGER NOT NULL DEFAULT 0, "callbackKey" TEXT UNIQUE, "callbackProcessedAt" DATETIME, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "SystemSetting" ("id" TEXT PRIMARY KEY NOT NULL, "key" TEXT NOT NULL UNIQUE, "value" TEXT NOT NULL, "category" TEXT NOT NULL DEFAULT 'general', "description" TEXT, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, "updatedAt" DATETIME NOT NULL);
`

func signAlipayValues(values url.Values, privateKey *rsa.PrivateKey) (string, error) {
	canonical := canonicalAlipayValues(values)
	digest := sha256.Sum256([]byte(canonical))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
