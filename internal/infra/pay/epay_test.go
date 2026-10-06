package pay

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestEpayChannelTypeFollowsServedMethod(t *testing.T) {
	alipay := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001", Key: "k"}, PaymentMethodAlipay, nil)
	if alipay.Method() != PaymentMethodAlipay || alipay.channelType() != "alipay" {
		t.Fatalf("alipay provider = %s / %s", alipay.Method(), alipay.channelType())
	}
	wechat := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001", Key: "k"}, PaymentMethodWeChat, nil)
	if wechat.Method() != PaymentMethodWeChat || wechat.channelType() != "wxpay" {
		t.Fatalf("wechat provider = %s / %s", wechat.Method(), wechat.channelType())
	}
}

func TestEpayConfiguredRequiresKeysPerVersion(t *testing.T) {
	base := EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001"}
	if EpayConfigured(base) {
		t.Fatal("v1 without merchant key must not be configured")
	}
	if !EpayConfigured(EpayConfiguration{APIURL: base.APIURL, PID: base.PID, Key: "merchant-key"}) {
		t.Fatal("v1 with merchant key should be configured")
	}
	if EpayConfigured(EpayConfiguration{APIURL: base.APIURL, PID: base.PID, Version: EpayVersionV2, MerchantPrivateKey: "private"}) {
		t.Fatal("v2 without platform public key must not be configured")
	}
	if !EpayConfigured(EpayConfiguration{APIURL: base.APIURL, PID: base.PID, Version: EpayVersionV2, MerchantPrivateKey: "private", PlatformPublicKey: "public"}) {
		t.Fatal("v2 with both keys should be configured")
	}
	if EpayConfigured(EpayConfiguration{PID: base.PID, Key: "k"}) {
		t.Fatal("missing gateway URL must not be configured")
	}
}

func TestCanonicalEpayValuesSortsAndSkipsBlankAndSignatureFields(t *testing.T) {
	values := url.Values{"type": {"alipay"}, "money": {"10.00"}, "name": {""}, "sign": {"forged"}, "sign_type": {"MD5"}, "out_trade_no": {"PAY-1"}}
	if got := canonicalEpayValues(values); got != "money=10.00&out_trade_no=PAY-1&type=alipay" {
		t.Fatalf("canonical = %q", got)
	}
}

func TestEpayV1CreateBuildsSignedSubmitURL(t *testing.T) {
	provider := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com/", PID: "1001", Key: "merchant-key", NotifyURL: "https://gw.example.com/notify", ReturnURL: "https://gw.example.com/return"}, PaymentMethodAlipay, nil)
	result, err := provider.Create(context.Background(), ProviderCreateRequest{OutTradeNo: "PAY-ABC", Amount: 10, Currency: "CNY", Subject: "积分充值", NotifyURL: "https://app.example.com/api/payment/callback/alipay", ReturnURL: "https://app.example.com/payment/result"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.PayURL, "https://epay.example.com/submit.php?") {
		t.Fatalf("pay URL = %q", result.PayURL)
	}
	parsed, err := url.Parse(result.PayURL)
	if err != nil {
		t.Fatal(err)
	}
	values := parsed.Query()
	if values.Get("pid") != "1001" || values.Get("type") != "alipay" || values.Get("out_trade_no") != "PAY-ABC" || values.Get("money") != "10.00" || values.Get("name") != "积分充值" || values.Get("notify_url") != "https://app.example.com/api/payment/callback/alipay" || values.Get("return_url") != "https://app.example.com/payment/result" || values.Get("sign_type") != "MD5" {
		t.Fatalf("submit values = %v", values)
	}
	// Recompute the signature from the documented algorithm instead of calling
	// the provider helper, so a broken canonicalization fails the test.
	signature := epayTestV1Signature(values, "merchant-key")
	if values.Get("sign") != signature {
		t.Fatalf("sign = %q, want %q", values.Get("sign"), signature)
	}
}

func epayTestV1Signature(values url.Values, key string) string {
	keys := make([]string, 0, len(values))
	for name, field := range values {
		if name == "sign" || name == "sign_type" || len(field) == 0 || field[0] == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+values.Get(name))
	}
	digest := md5.Sum([]byte(strings.Join(parts, "&") + key))
	return hex.EncodeToString(digest[:])
}

func TestEpayV1VerifyCallbackAcceptsValidAndRejectsTampering(t *testing.T) {
	provider := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001", Key: "merchant-key"}, PaymentMethodAlipay, nil)
	form := url.Values{
		"pid":          {"1001"},
		"trade_no":     {"2026080622555342651"},
		"out_trade_no": {"PAY-ABC"},
		"type":         {"alipay"},
		"name":         {"积分充值"},
		"money":        {"10.00"},
		"trade_status": {"TRADE_SUCCESS"},
		"sign_type":    {"MD5"},
	}
	form.Set("sign", epayTestV1Signature(form, "merchant-key"))
	callback, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form})
	if err != nil {
		t.Fatal(err)
	}
	if callback.OutTradeNo != "PAY-ABC" || callback.TransactionID != "2026080622555342651" || callback.Amount != 10 || callback.Currency != "CNY" || callback.Status != PaymentStatePaid || callback.CallbackKey != "epay:2026080622555342651" {
		t.Fatalf("callback = %+v", callback)
	}
	form.Set("money", "99.00")
	if _, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form}); err == nil {
		t.Fatal("expected tampered amount to break the signature")
	}
	form.Set("money", "10.00")
	form.Set("pid", "2002")
	form.Set("sign", epayTestV1Signature(form, "merchant-key"))
	if _, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form}); err == nil {
		t.Fatal("expected foreign pid to be rejected")
	}
	form.Set("pid", "1001")
	form.Set("trade_status", "WAIT_BUYER_PAY")
	form.Set("sign", epayTestV1Signature(form, "merchant-key"))
	callback, err = provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form})
	if err != nil {
		t.Fatal(err)
	}
	if callback.Status != PaymentStateFailed {
		t.Fatalf("non-success trade status = %s", callback.Status)
	}
}

func TestEpayV1QueryParsesPaidStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api.php" || r.URL.Query().Get("act") != "order" || r.URL.Query().Get("pid") != "1001" || r.URL.Query().Get("key") != "merchant-key" || r.URL.Query().Get("out_trade_no") != "PAY-ABC" {
			t.Errorf("unexpected query request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":1,"msg":"查询订单号成功！","trade_no":"T9","out_trade_no":"PAY-ABC","type":"alipay","status":1,"money":"1.00","endtime":"2026-08-06 22:56:12"}`)
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Key: "merchant-key"}, PaymentMethodAlipay, nil)
	status, err := provider.Query(context.Background(), "PAY-ABC")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != PaymentStatePaid || status.TransactionID != "T9" || status.Amount != 1 || status.Currency != "CNY" || status.PaidAt.IsZero() {
		t.Fatalf("status = %+v", status)
	}
}

func epayTestRSAKeypair(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}))
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}))
	return key, privatePEM, publicPEM
}

func epayTestV2Signature(values url.Values, privateKey *rsa.PrivateKey) string {
	keys := make([]string, 0, len(values))
	for name, field := range values {
		if name == "sign" || name == "sign_type" || len(field) == 0 || field[0] == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+values.Get(name))
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "&")))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}

func TestEpayV2CreateSendsSignedRequestAndParsesPayTarget(t *testing.T) {
	_, merchantPrivatePEM, merchantPublicPEM := epayTestRSAKeypair(t)
	var received url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pay/create" {
			t.Errorf("create path = %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse create form: %v", err)
		}
		received = r.Form
		publicKey, err := parseRSAPublicKey(merchantPublicPEM)
		if err != nil {
			t.Errorf("parse merchant public key: %v", err)
			return
		}
		if err := verifyRSASignature(publicKey, canonicalEpayValues(r.Form), r.Form.Get("sign")); err != nil {
			t.Errorf("create request signature: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":0,"trade_no":"T1","pay_type":"jump","pay_info":"https://pay.example.com/cashier/1"}`)
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Version: EpayVersionV2, MerchantPrivateKey: merchantPrivatePEM, PlatformPublicKey: merchantPublicPEM}, PaymentMethodWeChat, nil)
	result, err := provider.Create(context.Background(), ProviderCreateRequest{OutTradeNo: "PAY-ABC", Amount: 20, Currency: "CNY", Subject: "会员套餐", ClientIP: "203.0.113.7", NotifyURL: "https://app.example.com/api/payment/callback/wechat", ReturnURL: "https://app.example.com/payment/result"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PayURL != "https://pay.example.com/cashier/1" || result.QRCodeURL != "" {
		t.Fatalf("create result = %+v", result)
	}
	if received.Get("type") != "wxpay" || received.Get("method") != "jump" || received.Get("money") != "20.00" || received.Get("out_trade_no") != "PAY-ABC" || received.Get("clientip") != "203.0.113.7" || received.Get("timestamp") == "" {
		t.Fatalf("create request = %v", received)
	}
}

func TestEpayV2CreateMapsQrcodePayTypeToQRCodeURL(t *testing.T) {
	_, merchantPrivatePEM, merchantPublicPEM := epayTestRSAKeypair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":0,"trade_no":"T2","pay_type":"qrcode","pay_info":"weixin://wxpay/bizpayurl?pr=abc"}`)
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Version: EpayVersionV2, MerchantPrivateKey: merchantPrivatePEM, PlatformPublicKey: merchantPublicPEM}, PaymentMethodWeChat, nil)
	result, err := provider.Create(context.Background(), ProviderCreateRequest{OutTradeNo: "PAY-QR", Amount: 5, Currency: "CNY", Subject: "充值"})
	if err != nil {
		t.Fatal(err)
	}
	if result.QRCodeURL != "weixin://wxpay/bizpayurl?pr=abc" || result.PayURL != "" {
		t.Fatalf("create result = %+v", result)
	}
}

func TestEpayV2CreateRejectsGatewayError(t *testing.T) {
	_, merchantPrivatePEM, merchantPublicPEM := epayTestRSAKeypair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":-1,"msg":"签名错误"}`)
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Version: EpayVersionV2, MerchantPrivateKey: merchantPrivatePEM, PlatformPublicKey: merchantPublicPEM}, PaymentMethodAlipay, nil)
	if _, err := provider.Create(context.Background(), ProviderCreateRequest{OutTradeNo: "PAY-E", Amount: 1, Currency: "CNY", Subject: "x"}); err == nil || !strings.Contains(err.Error(), "签名错误") {
		t.Fatalf("expected gateway error, got %v", err)
	}
}

func TestEpayV2VerifyCallbackUsesPlatformKey(t *testing.T) {
	platformKey, _, platformPublicPEM := epayTestRSAKeypair(t)
	provider := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001", Version: EpayVersionV2, MerchantPrivateKey: "not-used", PlatformPublicKey: platformPublicPEM}, PaymentMethodAlipay, nil)
	form := url.Values{
		"pid":          {"1001"},
		"trade_no":     {"T3"},
		"out_trade_no": {"PAY-V2"},
		"type":         {"alipay"},
		"name":         {"会员"},
		"money":        {"30.00"},
		"trade_status": {"TRADE_SUCCESS"},
		"timestamp":    {time.Now().Format("2006-01-02 15:04:05")},
		"sign_type":    {"RSA"},
	}
	form.Set("sign", epayTestV2Signature(form, platformKey))
	callback, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form})
	if err != nil {
		t.Fatal(err)
	}
	if callback.Status != PaymentStatePaid || callback.Amount != 30 || callback.CallbackKey != "epay:T3" {
		t.Fatalf("callback = %+v", callback)
	}
	// Mutate the amount without re-signing: the signature still covers the
	// original value, so the gateway must reject the payload.
	form.Set("money", "999.00")
	if _, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form}); err == nil {
		t.Fatal("expected mutated V2 payload to be rejected")
	}
	// Signed with a foreign key must be rejected.
	_, foreignPrivatePEM, _ := epayTestRSAKeypair(t)
	foreignKey, err := parseRSAPrivateKey(foreignPrivatePEM)
	if err != nil {
		t.Fatal(err)
	}
	form.Set("money", "30.00")
	form.Set("sign", epayTestV2Signature(form, foreignKey))
	if _, err := provider.VerifyCallback(context.Background(), ProviderCallbackRequest{Form: form}); err == nil {
		t.Fatal("expected foreign platform key to be rejected")
	}
}

func TestEpayV2QueryAndCloseSendSignedRequests(t *testing.T) {
	_, merchantPrivatePEM, merchantPublicPEM := epayTestRSAKeypair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		publicKey, err := parseRSAPublicKey(merchantPublicPEM)
		if err != nil {
			t.Errorf("parse merchant public key: %v", err)
			return
		}
		if err := verifyRSASignature(publicKey, canonicalEpayValues(r.Form), r.Form.Get("sign")); err != nil {
			t.Errorf("%s signature: %v", r.URL.Path, err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/pay/query":
			_, _ = io.WriteString(w, `{"code":0,"trade_no":"T7","out_trade_no":"PAY-V2","status":1,"money":"5.00","endtime":"2026-08-06 22:56:12"}`)
		case "/api/pay/close":
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Version: EpayVersionV2, MerchantPrivateKey: merchantPrivatePEM, PlatformPublicKey: merchantPublicPEM}, PaymentMethodAlipay, nil)
	status, err := provider.Query(context.Background(), "PAY-V2")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != PaymentStatePaid || status.TransactionID != "T7" || status.Amount != 5 {
		t.Fatalf("status = %+v", status)
	}
	if err := provider.Cancel(context.Background(), "PAY-V2"); err != nil {
		t.Fatal(err)
	}
}

func TestEpayV1QueryReportsGatewayFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":-1,"msg":"商户ID不存在"}`)
	}))
	defer server.Close()
	provider := NewEpayProvider(EpayConfiguration{APIURL: server.URL, PID: "1001", Key: "merchant-key"}, PaymentMethodAlipay, nil)
	if _, err := provider.Query(context.Background(), "PAY-X"); err == nil || !strings.Contains(err.Error(), "商户ID不存在") {
		t.Fatalf("expected gateway failure, got %v", err)
	}
}

func TestEpayV1CancelIsLocalNoOp(t *testing.T) {
	provider := NewEpayProvider(EpayConfiguration{APIURL: "https://epay.example.com", PID: "1001", Key: "merchant-key"}, PaymentMethodAlipay, nil)
	if err := provider.Cancel(context.Background(), "PAY-X"); err != nil {
		t.Fatal(err)
	}
}
