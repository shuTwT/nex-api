package pay

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type PaymentMethod string

const (
	PaymentMethodWeChat PaymentMethod = "wechat"
	PaymentMethodAlipay PaymentMethod = "alipay"
	PaymentMethodMock   PaymentMethod = "mock"
)

type PaymentState string

const (
	PaymentStatePending   PaymentState = "pending"
	PaymentStatePaid      PaymentState = "paid"
	PaymentStateCancelled PaymentState = "cancelled"
	PaymentStateFailed    PaymentState = "failed"
	PaymentStateExpired   PaymentState = "expired"
)

var (
	ErrProviderUnavailable = errors.New("payment provider unavailable")
	ErrInvalidSignature    = errors.New("payment callback signature is invalid")
	ErrAmountMismatch      = errors.New("payment callback amount does not match order")
	ErrCurrencyMismatch    = errors.New("payment callback currency does not match order")
	ErrInvalidTransition   = errors.New("payment state transition is invalid")
	ErrPaymentAlreadyPaid  = errors.New("payment has already been paid")
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type ProviderCreateRequest struct {
	OutTradeNo string
	Amount     float64
	Currency   string
	NotifyURL  string
}

type ProviderCreateResult struct {
	QRCodeURL string
	PayURL    string
}

type ProviderCallbackRequest struct {
	Body    []byte
	Headers map[string][]string
	Form    map[string][]string
}

type ProviderCallback struct {
	OutTradeNo    string
	TransactionID string
	Amount        float64
	HasAmount     bool
	Currency      string
	Status        PaymentState
	PaidAt        time.Time
	CallbackKey   string
}

type ProviderStatus struct {
	TransactionID string
	Amount        float64
	Currency      string
	Status        PaymentState
	PaidAt        time.Time
}

type Provider interface {
	Method() PaymentMethod
	Create(ctx context.Context, request ProviderCreateRequest) (ProviderCreateResult, error)
	VerifyCallback(ctx context.Context, request ProviderCallbackRequest) (ProviderCallback, error)
	Query(ctx context.Context, outTradeNo string) (ProviderStatus, error)
	Cancel(ctx context.Context, outTradeNo string) error
}

func ComparePaymentAmount(expected float64, expectedCurrency string, actual float64, actualCurrency string) error {
	if strings.ToUpper(strings.TrimSpace(expectedCurrency)) != strings.ToUpper(strings.TrimSpace(actualCurrency)) {
		return ErrCurrencyMismatch
	}
	if cents(expected) != cents(actual) {
		return ErrAmountMismatch
	}
	return nil
}

func cents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

type transitionedPayment struct {
	Status      PaymentState
	PaidAt      time.Time
	CancelledAt time.Time
}

func TransitionPayment(current, target PaymentState, now time.Time) (transitionedPayment, error) {
	if current == target {
		return transitionedPayment{Status: current}, nil
	}
	if current != PaymentStatePending {
		return transitionedPayment{}, fmt.Errorf("%s to %s: %w", current, target, ErrInvalidTransition)
	}
	switch target {
	case PaymentStatePaid:
		return transitionedPayment{Status: target, PaidAt: now}, nil
	case PaymentStateCancelled:
		return transitionedPayment{Status: target, CancelledAt: now}, nil
	case PaymentStateFailed, PaymentStateExpired:
		return transitionedPayment{Status: target}, nil
	default:
		return transitionedPayment{}, fmt.Errorf("pending to %s: %w", target, ErrInvalidTransition)
	}
}

func parseRSAPublicKey(value string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("payment public key is not PEM")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if publicKey, ok := key.(*rsa.PublicKey); ok {
			return publicKey, nil
		}
	}
	if certificate, err := x509.ParseCertificate(block.Bytes); err == nil {
		if publicKey, ok := certificate.PublicKey.(*rsa.PublicKey); ok {
			return publicKey, nil
		}
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse payment public key: %w", err)
	}
	return key, nil
}

func parseRSAPrivateKey(value string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("payment private key is not PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if privateKey, ok := key.(*rsa.PrivateKey); ok {
			return privateKey, nil
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse payment private key: %w", err)
	}
	return key, nil
}

func verifyRSASignature(publicKey *rsa.PublicKey, message, encoded string) error {
	signature, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decode callback signature: %w", ErrInvalidSignature)
	}
	digest := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return fmt.Errorf("verify callback signature: %w", ErrInvalidSignature)
	}
	return nil
}
