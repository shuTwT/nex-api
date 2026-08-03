package pay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestVerifyWeChatSignatureRejectsForgedCallback(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)})
	err = VerifyWeChatSignature([]byte(`{"resource":{"ciphertext":"not-used"}}`), http.Header{"Wechatpay-Timestamp": {"1700000000"}, "Wechatpay-Nonce": {"nonce"}, "Wechatpay-Signature": {"forged"}}, string(publicKey))
	if err == nil {
		t.Fatal("expected forged signature to be rejected")
	}
}
func TestVerifyWeChatSignatureAcceptsPlatformCertificate(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{SerialNumber: big.NewInt(42), Subject: pkix.Name{CommonName: "WeChat Pay Platform"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	certificate, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"event_type":"TRANSACTION.SUCCESS"}`)
	message := "1700000000\nnonce\n" + string(body) + "\n"
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	headers := http.Header{"Wechatpay-Timestamp": {"1700000000"}, "Wechatpay-Nonce": {"nonce"}, "Wechatpay-Signature": {base64.StdEncoding.EncodeToString(signature)}}
	if err := VerifyWeChatSignature(body, headers, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}))); err != nil {
		t.Fatal(err)
	}
}
func TestWeChatCallbackSelectsPaymentPublicKeyBySerial(t *testing.T) {
	provider := NewWeChatProvider(WeChatConfiguration{PublicKey: "platform-certificate", PaymentPublicKey: "payment-public-key", PublicKeyID: "PUB_KEY_ID_1"}, nil)
	key, err := provider.callbackVerificationKey(http.Header{"Wechatpay-Serial": {"PUB_KEY_ID_1"}})
	if err != nil || key != "payment-public-key" {
		t.Fatalf("verification key = %q, %v", key, err)
	}
	key, err = provider.callbackVerificationKey(http.Header{"Wechatpay-Serial": {"CERT_SERIAL"}})
	if err != nil || key != "platform-certificate" {
		t.Fatalf("fallback key = %q, %v", key, err)
	}
}
func TestWeChatConfigurationAllowsPaymentPublicKeyMode(t *testing.T) {
	configuration := WeChatConfiguration{AppID: "app-id", MchID: "merchant-id", APIKey: "01234567890123456789012345678901", PrivateKey: "merchant-private-key", PaymentPublicKey: "payment-public-key", PublicKeyID: "PUB_KEY_ID_1"}
	if !WechatConfigured(configuration) {
		t.Fatal("payment public key mode should be configured")
	}
}
func TestVerifyAlipaySignatureAcceptsSignedPayloadAndRejectsMutation(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})
	values := url.Values{"out_trade_no": {"ALI-1"}, "trade_no": {"TX-1"}, "total_amount": {"10.00"}, "trade_status": {"TRADE_SUCCESS"}}
	signature, err := signAlipayValues(values, key)
	if err != nil {
		t.Fatal(err)
	}
	values.Set("sign", signature)
	if err := VerifyAlipaySignature(values, string(publicKey)); err != nil {
		t.Fatal(err)
	}
	values.Set("total_amount", "11.00")
	if err := VerifyAlipaySignature(values, string(publicKey)); err == nil {
		t.Fatal("expected mutated payload to be rejected")
	}
}
func TestComparePaymentAmountRequiresExactCentsAndCurrency(t *testing.T) {
	if err := ComparePaymentAmount(10, "CNY", 10, "CNY"); err != nil {
		t.Fatal(err)
	}
	if err := ComparePaymentAmount(10, "CNY", 10.01, "CNY"); err == nil {
		t.Fatal("expected amount mismatch")
	}
	if err := ComparePaymentAmount(10, "CNY", 10, "USD"); err == nil {
		t.Fatal("expected currency mismatch")
	}
}
func TestPaymentStateTransitionDoesNotOverridePaidOrCancelled(t *testing.T) {
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	paid, err := TransitionPayment(PaymentStatePending, PaymentStatePaid, now)
	if err != nil || paid.Status != PaymentStatePaid || paid.PaidAt != now {
		t.Fatalf("pending to paid transition failed: %#v %v", paid, err)
	}
	if _, err := TransitionPayment(PaymentStatePaid, PaymentStateCancelled, now); err == nil {
		t.Fatal("expected paid payment to reject cancellation")
	}
	if _, err := TransitionPayment(PaymentStateCancelled, PaymentStatePaid, now); err == nil {
		t.Fatal("expected cancelled payment to reject payment")
	}
}
func signAlipayValues(values url.Values, privateKey *rsa.PrivateKey) (string, error) {
	digest := sha256.Sum256([]byte(canonicalAlipayValues(values)))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
