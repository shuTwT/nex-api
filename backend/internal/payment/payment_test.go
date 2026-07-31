package payment

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/payment"
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

func TestServiceHandleCallbackGrantsRechargeOnceForDuplicateCallbacks(t *testing.T) {
	// Given
	client := newPaymentTestClient(t)
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	service, err := NewServiceWithProviders(client, config.Payment{}, []Provider{NewMockProvider(config.MockPayment{Enabled: true})}, fixedClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreatePayment(context.Background(), CreatePaymentInput{UserID: "u-1", Amount: 10, Currency: "CNY", Method: PaymentMethodMock, Metadata: map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage("10")}})
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

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

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
