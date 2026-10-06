// Package pay — epay.go implements the epay aggregator gateway in both
// protocol versions: V1 (MD5 signed submit.php/api.php, the Z-PAY dialect
// documented in docs/epay-integration.md) and V2 (RSA signed /api/pay/*).
package pay

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EpayVersion selects the epay protocol dialect of the gateway.
type EpayVersion string

const (
	// EpayVersionV1 is the classic MD5 protocol: submit.php redirect orders,
	// api.php?act=order queries, MD5 signed callbacks.
	EpayVersionV1 EpayVersion = "v1"
	// EpayVersionV2 is the RSA protocol: /api/pay/create orders, /api/pay/query
	// queries, SHA256withRSA signed callbacks with timestamp replay protection.
	EpayVersionV2 EpayVersion = "v2"
)

// EpayConfiguration holds the epay aggregator settings loaded from the system
// settings table by the payment service.
type EpayConfiguration struct {
	Enabled            bool
	APIURL             string
	PID                string
	Key                string // V1 merchant key (MD5 signing)
	Version            EpayVersion
	MerchantPrivateKey string // V2 merchant private key (PEM)
	PlatformPublicKey  string // V2 platform public key (PEM)
	NotifyURL          string
	ReturnURL          string
}

// EpayConfigured reports whether the epay provider can be built for the
// configured protocol version.
func EpayConfigured(value EpayConfiguration) bool {
	if !AllPresent(value.APIURL, value.PID) {
		return false
	}
	if value.Version == EpayVersionV2 {
		return AllPresent(value.MerchantPrivateKey, value.PlatformPublicKey)
	}
	return AllPresent(value.Key)
}

// EpayProvider routes one payment method through an epay aggregator gateway.
// The served method decides the epay channel type: alipay submits type=alipay
// and wechat submits type=wxpay, so both store-front entries keep working when
// the gateway replaces the direct channels.
type EpayProvider struct {
	config EpayConfiguration
	method PaymentMethod
	client *http.Client
}

// NewEpayProvider builds the gateway provider for one payment method.
func NewEpayProvider(configuration EpayConfiguration, method PaymentMethod, client *http.Client) *EpayProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if configuration.Version != EpayVersionV2 {
		configuration.Version = EpayVersionV1
	}
	return &EpayProvider{config: configuration, method: method, client: client}
}

func (p *EpayProvider) Method() PaymentMethod { return p.method }

// channelType maps the served payment method onto the epay type parameter.
func (p *EpayProvider) channelType() string {
	if p.method == PaymentMethodWeChat {
		return "wxpay"
	}
	return "alipay"
}

func (p *EpayProvider) baseURL() string {
	return strings.TrimRight(strings.TrimSpace(p.config.APIURL), "/")
}

// Create places an order on the epay gateway. V1 builds the signed submit.php
// redirect URL; V2 posts to /api/pay/create and unwraps the payment target
// from pay_info (pay_type=qrcode yields a QR code, everything else a pay URL).
func (p *EpayProvider) Create(ctx context.Context, request ProviderCreateRequest) (ProviderCreateResult, error) {
	subject := strings.TrimSpace(request.Subject)
	if subject == "" {
		subject = "支付订单"
	}
	notifyURL := firstNonEmpty(request.NotifyURL, p.config.NotifyURL)
	returnURL := firstNonEmpty(request.ReturnURL, p.config.ReturnURL)
	money := fmt.Sprintf("%.2f", request.Amount)
	if p.config.Version == EpayVersionV2 {
		return p.createV2(ctx, subject, money, notifyURL, returnURL, request)
	}
	return p.createV1(subject, money, notifyURL, returnURL, request)
}

func (p *EpayProvider) createV1(subject, money, notifyURL, returnURL string, request ProviderCreateRequest) (ProviderCreateResult, error) {
	values := url.Values{
		"pid":          {p.config.PID},
		"type":         {p.channelType()},
		"out_trade_no": {request.OutTradeNo},
		"notify_url":   {notifyURL},
		"return_url":   {returnURL},
		"name":         {subject},
		"money":        {money},
		"sign_type":    {"MD5"},
	}
	values.Set("sign", p.signV1(values))
	return ProviderCreateResult{PayURL: p.baseURL() + "/submit.php?" + values.Encode()}, nil
}

func (p *EpayProvider) createV2(ctx context.Context, subject, money, notifyURL, returnURL string, request ProviderCreateRequest) (ProviderCreateResult, error) {
	values := url.Values{
		"pid":          {p.config.PID},
		"method":       {"jump"},
		"type":         {p.channelType()},
		"out_trade_no": {request.OutTradeNo},
		"notify_url":   {notifyURL},
		"return_url":   {returnURL},
		"name":         {subject},
		"money":        {money},
		"timestamp":    {strconv.FormatInt(time.Now().Unix(), 10)},
		"sign_type":    {"RSA"},
	}
	// Several epay gateways reject the create order without the payer IP
	// (ezfp: 用户IP地址(clientip)不能为空). V1 submit.php does not take it.
	if clientIP := strings.TrimSpace(request.ClientIP); clientIP != "" {
		values.Set("clientip", clientIP)
	}
	signature, err := p.signV2(values)
	if err != nil {
		return ProviderCreateResult{}, err
	}
	values.Set("sign", signature)
	body, err := p.postForm(ctx, "/api/pay/create", values)
	if err != nil {
		return ProviderCreateResult{}, err
	}
	var payload struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		TradeNo string `json:"trade_no"`
		PayType string `json:"pay_type"`
		PayInfo string `json:"pay_info"`
		PayURL  string `json:"payurl"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProviderCreateResult{}, fmt.Errorf("decode epay create response: %w", err)
	}
	if payload.Code != 0 {
		return ProviderCreateResult{}, fmt.Errorf("epay create failed: %s", payload.Msg)
	}
	target := firstNonEmpty(payload.PayInfo, payload.PayURL, payload.URL)
	if target == "" {
		return ProviderCreateResult{}, errors.New("epay create response has no payment target")
	}
	if payload.PayType == "qrcode" {
		return ProviderCreateResult{QRCodeURL: target}, nil
	}
	return ProviderCreateResult{PayURL: target}, nil
}

// VerifyCallback validates the epay notify: V1 MD5 or V2 platform RSA
// signature over the sorted non-empty parameters, then the merchant pid and
// the trade status. It mirrors the alipay provider shape so the shared
// callback pipeline can consume it unchanged.
func (p *EpayProvider) VerifyCallback(_ context.Context, request ProviderCallbackRequest) (ProviderCallback, error) {
	form := request.Form
	if len(form) == 0 && len(request.Body) > 0 {
		parsed, err := url.ParseQuery(string(request.Body))
		if err != nil {
			return ProviderCallback{}, fmt.Errorf("parse epay callback form: %w", err)
		}
		form = parsed
	}
	if err := p.verifySignature(form); err != nil {
		return ProviderCallback{}, err
	}
	if firstHeader(form, "pid") != p.config.PID {
		return ProviderCallback{}, fmt.Errorf("epay callback pid mismatch: %w", ErrInvalidSignature)
	}
	status := PaymentStateFailed
	if firstHeader(form, "trade_status") == "TRADE_SUCCESS" {
		status = PaymentStatePaid
	}
	amount, err := strconv.ParseFloat(firstHeader(form, "money"), 64)
	if err != nil {
		return ProviderCallback{}, fmt.Errorf("parse epay callback amount: %w", err)
	}
	paidAt, _ := time.Parse("2006-01-02 15:04:05", firstHeader(form, "endtime"))
	transactionID := firstNonEmpty(firstHeader(form, "trade_no"), firstHeader(form, "out_trade_no"))
	return ProviderCallback{
		OutTradeNo:    firstHeader(form, "out_trade_no"),
		TransactionID: transactionID,
		Amount:        amount,
		HasAmount:     true,
		Currency:      "CNY",
		Status:        status,
		PaidAt:        paidAt,
		CallbackKey:   "epay:" + transactionID,
	}, nil
}

func (p *EpayProvider) verifySignature(form map[string][]string) error {
	signature := firstHeader(form, "sign")
	if signature == "" {
		return fmt.Errorf("epay callback without sign: %w", ErrInvalidSignature)
	}
	if p.config.Version == EpayVersionV2 {
		publicKey, err := parseRSAPublicKey(p.config.PlatformPublicKey)
		if err != nil {
			return fmt.Errorf("epay platform key: %w", err)
		}
		return verifyRSASignature(publicKey, canonicalEpayValues(form), signature)
	}
	expected := epayV1Signature(form, p.config.Key)
	if !strings.EqualFold(expected, signature) {
		return fmt.Errorf("verify epay callback: %w", ErrInvalidSignature)
	}
	return nil
}

// Query fetches the order state from the gateway: V1 GET api.php?act=order,
// V2 signed POST /api/pay/query. Only paid state is reported as final; every
// other code keeps the order pending so the local expiration job stays the
// source of truth for abandoned orders.
func (p *EpayProvider) Query(ctx context.Context, outTradeNo string) (ProviderStatus, error) {
	if p.config.Version == EpayVersionV2 {
		return p.queryV2(ctx, outTradeNo)
	}
	return p.queryV1(ctx, outTradeNo)
}

func (p *EpayProvider) queryV1(ctx context.Context, outTradeNo string) (ProviderStatus, error) {
	endpoint := p.baseURL() + "/api.php?" + url.Values{
		"act":          {"order"},
		"pid":          {p.config.PID},
		"key":          {p.config.Key},
		"out_trade_no": {outTradeNo},
	}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("create epay query request: %w", err)
	}
	body, err := p.do(request)
	if err != nil {
		return ProviderStatus{}, err
	}
	var payload struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		TradeNo string `json:"trade_no"`
		Status  int    `json:"status"`
		Money   string `json:"money"`
		EndTime string `json:"endtime"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProviderStatus{}, fmt.Errorf("decode epay query response: %w", err)
	}
	if payload.Code != 1 {
		return ProviderStatus{}, fmt.Errorf("epay query failed: %s", payload.Msg)
	}
	status := PaymentStatePending
	if payload.Status == 1 {
		status = PaymentStatePaid
	}
	amount, err := strconv.ParseFloat(payload.Money, 64)
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("parse epay query amount: %w", err)
	}
	paidAt, _ := time.Parse("2006-01-02 15:04:05", payload.EndTime)
	return ProviderStatus{TransactionID: payload.TradeNo, Amount: amount, Currency: "CNY", Status: status, PaidAt: paidAt}, nil
}

func (p *EpayProvider) queryV2(ctx context.Context, outTradeNo string) (ProviderStatus, error) {
	values := url.Values{
		"pid":          {p.config.PID},
		"out_trade_no": {outTradeNo},
		"timestamp":    {strconv.FormatInt(time.Now().Unix(), 10)},
		"sign_type":    {"RSA"},
	}
	signature, err := p.signV2(values)
	if err != nil {
		return ProviderStatus{}, err
	}
	values.Set("sign", signature)
	body, err := p.postForm(ctx, "/api/pay/query", values)
	if err != nil {
		return ProviderStatus{}, err
	}
	var payload struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		TradeNo string `json:"trade_no"`
		Status  int    `json:"status"`
		Money   string `json:"money"`
		EndTime string `json:"endtime"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProviderStatus{}, fmt.Errorf("decode epay query response: %w", err)
	}
	if payload.Code != 0 {
		return ProviderStatus{}, fmt.Errorf("epay query failed: %s", payload.Msg)
	}
	status := PaymentStatePending
	if payload.Status == 1 || payload.Status == 2 {
		status = PaymentStatePaid
	}
	amount, err := strconv.ParseFloat(payload.Money, 64)
	if err != nil {
		return ProviderStatus{}, fmt.Errorf("parse epay query amount: %w", err)
	}
	paidAt, _ := time.Parse("2006-01-02 15:04:05", payload.EndTime)
	return ProviderStatus{TransactionID: payload.TradeNo, Amount: amount, Currency: "CNY", Status: status, PaidAt: paidAt}, nil
}

// Cancel is a no-op on V1 (no close endpoint; pending orders expire locally)
// and closes the gateway order on V2.
func (p *EpayProvider) Cancel(ctx context.Context, outTradeNo string) error {
	if p.config.Version != EpayVersionV2 {
		return nil
	}
	values := url.Values{
		"pid":          {p.config.PID},
		"out_trade_no": {outTradeNo},
		"timestamp":    {strconv.FormatInt(time.Now().Unix(), 10)},
		"sign_type":    {"RSA"},
	}
	signature, err := p.signV2(values)
	if err != nil {
		return err
	}
	values.Set("sign", signature)
	body, err := p.postForm(ctx, "/api/pay/close", values)
	if err != nil {
		return err
	}
	var payload struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode epay close response: %w", err)
	}
	if payload.Code != 0 {
		return fmt.Errorf("epay close failed: %s", payload.Msg)
	}
	return nil
}

func (p *EpayProvider) postForm(ctx context.Context, path string, values url.Values) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL()+path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create epay request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return p.do(request)
}

func (p *EpayProvider) do(request *http.Request) ([]byte, error) {
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("epay request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read epay response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("epay request returned status %d", response.StatusCode)
	}
	return body, nil
}

// signV1 returns the V1 MD5 signature: md5(sorted k=v string + merchant key),
// lowercase hex.
func (p *EpayProvider) signV1(values url.Values) string {
	return epayV1Signature(values, p.config.Key)
}

func epayV1Signature(values map[string][]string, key string) string {
	digest := md5.Sum([]byte(canonicalEpayValues(values) + key))
	return hex.EncodeToString(digest[:])
}

// signV2 returns the V2 signature: SHA256withRSA over the canonical string
// with the merchant private key, base64 encoded.
func (p *EpayProvider) signV2(values url.Values) (string, error) {
	privateKey, err := parseRSAPrivateKey(p.config.MerchantPrivateKey)
	if err != nil {
		return "", fmt.Errorf("epay merchant key: %w", err)
	}
	digest := sha256.Sum256([]byte(canonicalEpayValues(values)))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign epay request: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// canonicalEpayValues joins the non-empty parameters (excluding sign and
// sign_type) sorted by key as k=v&k=v — the signing input shared by both
// protocol versions. It accepts url.Values and callback form maps alike.
func canonicalEpayValues(values map[string][]string) string {
	keys := make([]string, 0, len(values))
	for key, field := range values {
		if key == "sign" || key == "sign_type" || len(field) == 0 || strings.TrimSpace(field[0]) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+firstHeader(values, key))
	}
	return strings.Join(parts, "&")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
