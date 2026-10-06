package payment

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/enttest"
	"github.com/shuTwT/nex-api/ent/payment"
	"github.com/shuTwT/nex-api/ent/systemsetting"
	pay "github.com/shuTwT/nex-api/internal/infra/pay"
)

func TestServiceCreatePaymentReadsCurrentNotifyURLFromDatabase(t *testing.T) {
	client := newPaymentTestClient(t)
	for key, value := range map[string]string{"wechatPayAppId": "app-id", "wechatPayMchId": "merchant-id", "wechatPayApiKey": "01234567890123456789012345678901", "wechatPayPrivateKey": "private-key", "wechatPayPublicKey": "public-key", "wechatPayNotifyUrl": "https://example.com/first-callback", "wechatEnabled": "true"} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	provider := &recordingProvider{method: pay.PaymentMethodWeChat}
	service.providerFactory = func(pay.PaymentMethod, pay.PaymentConfiguration) pay.Provider { return provider }
	if _, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: pay.PaymentMethodWeChat, ClientIP: "203.0.113.9"}); err != nil {
		t.Fatal(err)
	}
	if provider.createRequest.NotifyURL != "https://example.com/first-callback" {
		t.Fatalf("notify URL = %q", provider.createRequest.NotifyURL)
	}
	if provider.createRequest.ClientIP != "203.0.113.9" {
		t.Fatalf("client IP = %q", provider.createRequest.ClientIP)
	}
	setting, err := client.SystemSetting.Query().Where(systemsetting.Key("wechatPayNotifyUrl")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setting.Update().SetValue("https://example.com/updated-callback").Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 20, Currency: "CNY", Method: pay.PaymentMethodWeChat}); err != nil {
		t.Fatal(err)
	}
	if provider.createRequest.NotifyURL != "https://example.com/updated-callback" {
		t.Fatalf("updated notify URL = %q", provider.createRequest.NotifyURL)
	}
}

func TestPaymentConfigurationUsesAppURLFallbacks(t *testing.T) {
	service, err := NewService(newPaymentTestClient(t), "https://app.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	configuration, err := service.loadConfiguration(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if configuration.WeChat.NotifyURL != "https://app.example.com/api/payment/callback/wechat" || configuration.Alipay.NotifyURL != "https://app.example.com/api/payment/callback/alipay" || configuration.Alipay.ReturnURL != "https://app.example.com/payment/result" {
		t.Fatalf("configuration = %+v", configuration)
	}
}
func TestMockPaymentAutoSuccessAndStatusQuery(t *testing.T) {
	client := newPaymentTestClient(t)
	for key, value := range map[string]string{"mockPaymentEnabled": "true", "mockPaymentAutoSuccess": "true", "mockPaymentDelay": "25"} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	var delay time.Duration
	service.schedule = func(got time.Duration, callback func()) { delay = got; callback() }
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: pay.PaymentMethodMock, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")}})
	if err != nil {
		t.Fatal(err)
	}
	if created.PayURL != "/payment/mock?outTradeNo="+url.QueryEscape(created.OutTradeNo) {
		t.Fatalf("pay URL = %q", created.PayURL)
	}
	if delay != 25*time.Millisecond {
		t.Fatalf("delay = %s", delay)
	}
	view, err := service.QueryPayment(context.Background(), created.OutTradeNo)
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != pay.PaymentStatePaid {
		t.Fatalf("status = %s", view.Status)
	}
	user, err := client.User.Get(context.Background(), "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 1010 {
		t.Fatalf("credits = %d", user.Credits)
	}
}
func TestMockPaymentFailureUsesFailedState(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewServiceWithProviders(client, []pay.Provider{pay.NewMockProvider(pay.MockPaymentConfiguration{Enabled: true})})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: pay.PaymentMethodMock})
	if err != nil {
		t.Fatal(err)
	}
	view, err := service.HandleCallback(context.Background(), pay.PaymentMethodMock, pay.ProviderCallbackRequest{Body: []byte(`{"outTradeNo":"` + created.OutTradeNo + `","success":false}`)})
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != pay.PaymentStateFailed {
		t.Fatalf("status = %s", view.Status)
	}
}
func TestServiceHandleCallbackGrantsRechargeOnceForDuplicateCallbacks(t *testing.T) {
	client := newPaymentTestClient(t)
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	service, err := NewServiceWithProviders(client, []pay.Provider{pay.NewMockProvider(pay.MockPaymentConfiguration{Enabled: true})}, fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: pay.PaymentMethodMock, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")}})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"outTradeNo":"` + created.OutTradeNo + `","success":true,"amount":10,"currency":"CNY"}`)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := service.HandleCallback(context.Background(), pay.PaymentMethodMock, pay.ProviderCallbackRequest{Body: body}); err != nil {
				t.Errorf("callback: %v", err)
			}
		}()
	}
	wg.Wait()
	user, err := client.User.Get(context.Background(), "u-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 1010 {
		t.Fatalf("credits = %d", user.Credits)
	}
	record, err := client.Payment.Query().Where(payment.OutTradeNo(created.OutTradeNo)).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != string(pay.PaymentStatePaid) || record.CallbackVersion != 1 {
		t.Fatalf("record = %+v", record)
	}
}
func TestExpirePendingPaymentsClosesProviderAndMarksOrderExpired(t *testing.T) {
	client := newPaymentTestClient(t)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	provider := &recordingProvider{method: pay.PaymentMethodWeChat}
	service, err := NewServiceWithProviders(client, []pay.Provider{provider}, fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	order, err := client.Payment.Create().SetUserId("u-1").SetOutTradeNo("PAY-EXPIRED").SetMethod(string(pay.PaymentMethodWeChat)).SetAmount(10).SetCurrency("CNY").SetStatus(string(pay.PaymentStatePending)).SetExpiredAt(now.Add(-time.Minute)).SetCreatedAt(now.Add(-3 * time.Hour)).SetUpdatedAt(now.Add(-3 * time.Hour)).Save(context.Background())
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
	if updated.Status != string(pay.PaymentStateExpired) || updated.CallbackVersion != 1 {
		t.Fatalf("record = %+v", updated)
	}
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type recordingProvider struct {
	method        pay.PaymentMethod
	createRequest pay.ProviderCreateRequest
	cancelled     []string
}

func (p *recordingProvider) Method() pay.PaymentMethod { return p.method }
func (p *recordingProvider) Create(_ context.Context, r pay.ProviderCreateRequest) (pay.ProviderCreateResult, error) {
	p.createRequest = r
	return pay.ProviderCreateResult{PayURL: "https://pay.example.com"}, nil
}
func (p *recordingProvider) VerifyCallback(context.Context, pay.ProviderCallbackRequest) (pay.ProviderCallback, error) {
	return pay.ProviderCallback{}, nil
}
func (p *recordingProvider) Query(context.Context, string) (pay.ProviderStatus, error) {
	return pay.ProviderStatus{Status: pay.PaymentStatePending}, nil
}
func (p *recordingProvider) Cancel(_ context.Context, outTradeNo string) error {
	p.cancelled = append(p.cancelled, outTradeNo)
	return nil
}
func newPaymentTestClient(t *testing.T) *ent.Client {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/payment.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	if _, err := client.User.Create().SetID("u-1").SetName("User").SetEmail("u-1@example.com").SetUsername("u-1").SetPassword("redacted").SetRole("user").SetCredits(1000).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	return client
}

func createEpaySettings(t *testing.T, client *ent.Client, version string) {
	t.Helper()
	for key, value := range map[string]string{"epayEnabled": "true", "epayApiUrl": "https://epay.example.com", "epayPid": "1001", "epayKey": "merchant-key", "epayVersion": version} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func containsMethod(methods []pay.PaymentMethod, target pay.PaymentMethod) bool {
	for _, method := range methods {
		if method == target {
			return true
		}
	}
	return false
}

func TestEpayGatewayServesAlipayAndWechatMethodsWhenEnabled(t *testing.T) {
	client := newPaymentTestClient(t)
	createEpaySettings(t, client, "v1")
	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	methods, err := service.AvailableMethods(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !containsMethod(methods, pay.PaymentMethodAlipay) || !containsMethod(methods, pay.PaymentMethodWeChat) {
		t.Fatalf("methods = %v, want alipay and wechat served by epay", methods)
	}
	settings, err := service.Settings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !settings.AlipayEnabled || !settings.WeChatEnabled {
		t.Fatalf("settings = %+v, want both channels reported available via epay", settings)
	}
	provider, notifyURL, err := service.resolveProvider(context.Background(), pay.PaymentMethodAlipay)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*pay.EpayProvider); !ok {
		t.Fatalf("alipay provider = %T, want *pay.EpayProvider", provider)
	}
	if notifyURL != "https://app.example.com/api/payment/callback/alipay" {
		t.Fatalf("notify URL = %q", notifyURL)
	}
	provider, notifyURL, err = service.resolveProvider(context.Background(), pay.PaymentMethodWeChat)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := provider.(*pay.EpayProvider); !ok {
		t.Fatalf("wechat provider = %T, want *pay.EpayProvider", provider)
	}
	if notifyURL != "https://app.example.com/api/payment/callback/wechat" {
		t.Fatalf("notify URL = %q", notifyURL)
	}
}

func TestDisabledDirectChannelsHiddenWhenEpayOff(t *testing.T) {
	client := newPaymentTestClient(t)
	for key, value := range map[string]string{"alipayAppId": "app-id", "alipayPrivateKey": "private-key", "alipayPublicKey": "public-key"} {
		if _, err := client.SystemSetting.Create().SetKey(key).SetValue(value).SetCategory("payment").Save(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	service, err := NewService(client, "https://app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	methods, err := service.AvailableMethods(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if containsMethod(methods, pay.PaymentMethodAlipay) {
		t.Fatalf("methods = %v, alipay is configured but disabled", methods)
	}
	if _, _, err := service.resolveProvider(context.Background(), pay.PaymentMethodAlipay); err == nil {
		t.Fatal("expected disabled alipay to be rejected")
	}
}

func TestEpayOrderUsesGatewayMetadataAndShortTradeNo(t *testing.T) {
	client := newPaymentTestClient(t)
	service, err := NewServiceWithProviders(client, []pay.Provider{pay.NewEpayProvider(pay.EpayConfiguration{Enabled: true, APIURL: "https://epay.example.com", PID: "1001", Key: "merchant-key", Version: pay.EpayVersionV1, ReturnURL: "https://app.example.com/payment/result"}, pay.PaymentMethodAlipay, nil)})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.createPayment(context.Background(), createPaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: pay.PaymentMethodAlipay, Subject: "积分充值", Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.OutTradeNo) != 32 || !strings.HasPrefix(created.OutTradeNo, "PAY-") {
		t.Fatalf("out trade no = %q, must be PAY- prefixed and within 32 characters", created.OutTradeNo)
	}
	if !strings.HasPrefix(created.PayURL, "https://epay.example.com/submit.php?") {
		t.Fatalf("pay URL = %q", created.PayURL)
	}
	record, err := client.Payment.Query().Where(payment.OutTradeNo(created.OutTradeNo)).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(record.Metadata), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["gateway"] != "epay" || metadata["type"] != "recharge" || metadata["credits"] != float64(10) {
		t.Fatalf("metadata = %v", metadata)
	}
}
