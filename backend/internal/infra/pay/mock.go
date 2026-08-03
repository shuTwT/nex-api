package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type MockProvider struct {
	config MockPaymentConfiguration
}

func NewMockProvider(paymentConfig MockPaymentConfiguration) *MockProvider {
	return &MockProvider{config: paymentConfig}
}

func (p *MockProvider) Method() PaymentMethod { return PaymentMethodMock }

// AutoSuccess reports whether mock payments auto-succeed after the configured
// delay; used by the payment service to schedule callback simulation.
func (p *MockProvider) AutoSuccess() bool { return p.config.AutoSuccess }

// Delay returns the mock auto-success delay.
func (p *MockProvider) Delay() time.Duration { return p.config.Delay }

func (p *MockProvider) Create(_ context.Context, request ProviderCreateRequest) (ProviderCreateResult, error) {
	if !p.config.Enabled {
		return ProviderCreateResult{}, ErrProviderUnavailable
	}
	return ProviderCreateResult{PayURL: "/payment/mock?outTradeNo=" + url.QueryEscape(request.OutTradeNo)}, nil
}

func (p *MockProvider) VerifyCallback(_ context.Context, request ProviderCallbackRequest) (ProviderCallback, error) {
	var payload struct {
		OutTradeNo    string  `json:"outTradeNo"`
		Success       *bool   `json:"success"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		TransactionID string  `json:"transactionId"`
	}
	if err := json.Unmarshal(request.Body, &payload); err != nil {
		return ProviderCallback{}, fmt.Errorf("decode mock callback: %w", err)
	}
	if strings.TrimSpace(payload.OutTradeNo) == "" {
		return ProviderCallback{}, errors.New("mock callback order number is required")
	}
	success := true
	if payload.Success != nil {
		success = *payload.Success
	}
	status := PaymentStateFailed
	if success {
		status = PaymentStatePaid
	}
	if payload.TransactionID == "" {
		payload.TransactionID = "MOCK-" + payload.OutTradeNo
	}
	callback := ProviderCallback{
		OutTradeNo: payload.OutTradeNo, TransactionID: payload.TransactionID,
		Amount: payload.Amount, HasAmount: payload.Amount > 0, Currency: payload.Currency,
		Status: status, CallbackKey: "mock:" + payload.OutTradeNo + ":" + string(status),
	}
	if status == PaymentStatePaid {
		callback.PaidAt = time.Now().UTC()
	}
	return callback, nil
}

func (p *MockProvider) Query(_ context.Context, _ string) (ProviderStatus, error) {
	return ProviderStatus{}, errors.New("mock provider has no remote query")
}

func (p *MockProvider) Cancel(_ context.Context, _ string) error {
	if !p.config.Enabled {
		return ErrProviderUnavailable
	}
	return nil
}
