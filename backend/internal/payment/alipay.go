package payment

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AlipayProvider struct {
	config alipayConfiguration
	client *http.Client
}

func NewAlipayProvider(paymentConfig alipayConfiguration, client *http.Client) *AlipayProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &AlipayProvider{config: paymentConfig, client: client}
}

func (p *AlipayProvider) Method() PaymentMethod { return PaymentMethodAlipay }

func VerifyAlipaySignature(values url.Values, publicKeyPEM string) error {
	publicKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("alipay callback key: %w", err)
	}
	return verifyRSASignature(publicKey, canonicalAlipayValues(values), values.Get("sign"))
}

func canonicalAlipayValues(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "sign" && key != "sign_type" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}
	return strings.Join(parts, "&")
}

func (p *AlipayProvider) VerifyCallback(_ context.Context, request ProviderCallbackRequest) (ProviderCallback, error) {
	if err := VerifyAlipaySignature(request.Form, p.config.AlipayPublicKey); err != nil {
		return ProviderCallback{}, err
	}
	amount, err := strconv.ParseFloat(request.Form.Get("total_amount"), 64)
	if err != nil {
		return ProviderCallback{}, fmt.Errorf("parse alipay amount: %w", err)
	}
	status := PaymentStateFailed
	if request.Form.Get("trade_status") == "TRADE_SUCCESS" || request.Form.Get("trade_status") == "TRADE_FINISHED" {
		status = PaymentStatePaid
	}
	paidAt, _ := time.Parse("2006-01-02 15:04:05", request.Form.Get("gmt_payment"))
	transactionID := request.Form.Get("trade_no")
	return ProviderCallback{
		OutTradeNo: request.Form.Get("out_trade_no"), TransactionID: transactionID,
		Amount: amount, HasAmount: true, Currency: "CNY", Status: status, PaidAt: paidAt,
		CallbackKey: "alipay:" + transactionID,
	}, nil
}

func (p *AlipayProvider) Create(ctx context.Context, request ProviderCreateRequest) (ProviderCreateResult, error) {
	bizContent, err := json.Marshal(struct {
		OutTradeNo  string `json:"out_trade_no"`
		ProductCode string `json:"product_code"`
		TotalAmount string `json:"total_amount"`
		Subject     string `json:"subject"`
	}{OutTradeNo: request.OutTradeNo, ProductCode: "FAST_INSTANT_TRADE_PAY", TotalAmount: fmt.Sprintf("%.2f", request.Amount), Subject: "支付订单"})
	if err != nil {
		return ProviderCreateResult{}, fmt.Errorf("encode alipay payment: %w", err)
	}
	values := url.Values{
		"app_id":      {p.config.AppID},
		"method":      {"alipay.trade.page.pay"},
		"format":      {"JSON"},
		"charset":     {"utf-8"},
		"sign_type":   {"RSA2"},
		"timestamp":   {time.Now().Format("2006-01-02 15:04:05")},
		"version":     {"1.0"},
		"notify_url":  {request.NotifyURL},
		"return_url":  {p.config.ReturnURL},
		"biz_content": {string(bizContent)},
	}
	signature, err := p.sign(values)
	if err != nil {
		return ProviderCreateResult{}, err
	}
	values.Set("sign", signature)
	gateway := "https://openapi.alipay.com/gateway.do"
	if p.config.Sandbox {
		gateway = "https://openapi.alipaydev.com/gateway.do"
	}
	return ProviderCreateResult{PayURL: gateway + "?" + values.Encode()}, nil
}

func (p *AlipayProvider) Query(ctx context.Context, outTradeNo string) (ProviderStatus, error) {
	content, err := json.Marshal(struct {
		OutTradeNo string `json:"out_trade_no"`
	}{OutTradeNo: outTradeNo})
	if err != nil {
		return ProviderStatus{}, err
	}
	result, err := p.execute(ctx, "alipay.trade.query", content)
	if err != nil {
		return ProviderStatus{}, err
	}
	var payload struct {
		TradeQuery struct {
			TradeStatus string `json:"trade_status"`
			TradeNo     string `json:"trade_no"`
			TotalAmount string `json:"total_amount"`
			SendPayDate string `json:"send_pay_date"`
		} `json:"alipay_trade_query_response"`
	}
	if err := json.Unmarshal(result, &payload); err != nil {
		return ProviderStatus{}, fmt.Errorf("decode alipay query response: %w", err)
	}
	status := PaymentStatePending
	if payload.TradeQuery.TradeStatus == "TRADE_SUCCESS" || payload.TradeQuery.TradeStatus == "TRADE_FINISHED" {
		status = PaymentStatePaid
	} else if payload.TradeQuery.TradeStatus == "TRADE_CLOSED" {
		status = PaymentStateCancelled
	}
	amount, err := strconv.ParseFloat(payload.TradeQuery.TotalAmount, 64)
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("parse alipay query amount: %w", err)
	}
	paidAt, _ := time.Parse("2006-01-02 15:04:05", payload.TradeQuery.SendPayDate)
	return ProviderStatus{TransactionID: payload.TradeQuery.TradeNo, Amount: amount, Currency: "CNY", Status: status, PaidAt: paidAt}, nil
}

func (p *AlipayProvider) Cancel(ctx context.Context, outTradeNo string) error {
	content, err := json.Marshal(struct {
		OutTradeNo string `json:"out_trade_no"`
	}{OutTradeNo: outTradeNo})
	if err != nil {
		return err
	}
	_, err = p.execute(ctx, "alipay.trade.close", content)
	return err
}

func (p *AlipayProvider) sign(values url.Values) (string, error) {
	privateKey, err := parseRSAPrivateKey(p.config.PrivateKey)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(canonicalAlipayValues(values)))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign alipay request: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (p *AlipayProvider) execute(ctx context.Context, method string, bizContent []byte) ([]byte, error) {
	values := url.Values{"app_id": {p.config.AppID}, "method": {method}, "format": {"JSON"}, "charset": {"utf-8"}, "sign_type": {"RSA2"}, "timestamp": {time.Now().Format("2006-01-02 15:04:05")}, "version": {"1.0"}, "biz_content": {string(bizContent)}}
	signature, err := p.sign(values)
	if err != nil {
		return nil, err
	}
	values.Set("sign", signature)
	gateway := "https://openapi.alipay.com/gateway.do"
	if p.config.Sandbox {
		gateway = "https://openapi.alipaydev.com/gateway.do"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, gateway, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create alipay request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("alipay request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read alipay response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("alipay request returned status %d", response.StatusCode)
	}
	return body, nil
}
