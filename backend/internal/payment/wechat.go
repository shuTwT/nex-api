package payment

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
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
	"strconv"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/config"
)

type WeChatProvider struct {
	config config.WeChatPayment
	client *http.Client
}

func NewWeChatProvider(paymentConfig config.WeChatPayment, client *http.Client) *WeChatProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &WeChatProvider{config: paymentConfig, client: client}
}

func (p *WeChatProvider) Method() PaymentMethod { return PaymentMethodWeChat }

func VerifyWeChatSignature(body []byte, headers http.Header, publicKeyPEM string) error {
	timestamp := headers.Get("Wechatpay-Timestamp")
	nonce := headers.Get("Wechatpay-Nonce")
	signature := headers.Get("Wechatpay-Signature")
	if timestamp == "" || nonce == "" || signature == "" {
		return ErrInvalidSignature
	}
	publicKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("wechat callback key: %w", err)
	}
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	return verifyRSASignature(publicKey, message, signature)
}

func (p *WeChatProvider) VerifyCallback(_ context.Context, request ProviderCallbackRequest) (ProviderCallback, error) {
	if err := VerifyWeChatSignature(request.Body, request.Headers, p.config.PublicKey); err != nil {
		return ProviderCallback{}, err
	}
	var envelope wechatCallbackEnvelope
	if err := json.Unmarshal(request.Body, &envelope); err != nil {
		return ProviderCallback{}, fmt.Errorf("decode wechat callback: %w", err)
	}
	resource, err := p.decrypt(envelope.Resource)
	if err != nil {
		return ProviderCallback{}, err
	}
	paidAt, err := time.Parse(time.RFC3339, resource.SuccessTime)
	if err != nil {
		return ProviderCallback{}, fmt.Errorf("parse wechat success time: %w", err)
	}
	return ProviderCallback{
		OutTradeNo:    resource.OutTradeNo,
		TransactionID: resource.TransactionID,
		Amount:        float64(resource.Amount.Total) / 100,
		HasAmount:     true,
		Currency:      resource.Amount.Currency,
		Status:        PaymentStatePaid,
		PaidAt:        paidAt,
		CallbackKey:   "wechat:" + resource.TransactionID,
	}, nil
}

func (p *WeChatProvider) Create(ctx context.Context, request ProviderCreateRequest) (ProviderCreateResult, error) {
	body, err := json.Marshal(struct {
		AppID       string `json:"appid"`
		MchID       string `json:"mchid"`
		Description string `json:"description"`
		OutTradeNo  string `json:"out_trade_no"`
		NotifyURL   string `json:"notify_url"`
		Amount      struct {
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
	}{
		AppID: p.config.AppID, MchID: p.config.MchID, Description: "支付订单",
		OutTradeNo: request.OutTradeNo, NotifyURL: request.NotifyURL,
		Amount: struct {
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		}{Total: cents(request.Amount), Currency: request.Currency},
	})
	if err != nil {
		return ProviderCreateResult{}, fmt.Errorf("encode wechat payment: %w", err)
	}
	response, err := p.doSigned(ctx, http.MethodPost, "/v3/pay/transactions/native", body)
	if err != nil {
		return ProviderCreateResult{}, err
	}
	var result struct {
		CodeURL string `json:"code_url"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return ProviderCreateResult{}, fmt.Errorf("decode wechat create response: %w", err)
	}
	if result.CodeURL == "" {
		return ProviderCreateResult{}, errors.New("wechat payment response has no code url")
	}
	return ProviderCreateResult{QRCodeURL: result.CodeURL}, nil
}

func (p *WeChatProvider) Query(ctx context.Context, outTradeNo string) (ProviderStatus, error) {
	path := "/v3/pay/transactions/out-trade-no/" + url.PathEscape(outTradeNo) + "?mchid=" + url.QueryEscape(p.config.MchID)
	response, err := p.doSigned(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ProviderStatus{}, err
	}
	var result wechatQueryResponse
	if err := json.Unmarshal(response, &result); err != nil {
		return ProviderStatus{}, fmt.Errorf("decode wechat query response: %w", err)
	}
	status := PaymentStatePending
	switch result.TradeState {
	case "SUCCESS":
		status = PaymentStatePaid
	case "CLOSED", "PAYERROR":
		status = PaymentStateCancelled
	}
	paidAt, _ := time.Parse(time.RFC3339, result.SuccessTime)
	return ProviderStatus{TransactionID: result.TransactionID, Amount: float64(result.Amount.Total) / 100, Currency: result.Amount.Currency, Status: status, PaidAt: paidAt}, nil
}

func (p *WeChatProvider) Cancel(ctx context.Context, outTradeNo string) error {
	path := "/v3/pay/transactions/out-trade-no/" + url.PathEscape(outTradeNo) + "/close"
	body := []byte(`{"mchid":"` + p.config.MchID + `"}`)
	_, err := p.doSigned(ctx, http.MethodPost, path, body)
	return err
}

type wechatCallbackEnvelope struct {
	Resource wechatEncryptedResource `json:"resource"`
}

type wechatEncryptedResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type wechatPaymentResource struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	SuccessTime   string `json:"success_time"`
	Amount        struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

type wechatQueryResponse struct {
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	SuccessTime   string `json:"success_time"`
	Amount        struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

func (p *WeChatProvider) decrypt(resource wechatEncryptedResource) (wechatPaymentResource, error) {
	key := []byte(p.config.APIKey)
	if len(key) != 32 {
		return wechatPaymentResource{}, errors.New("wechat api key must be 32 bytes")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(resource.Ciphertext)
	if err != nil {
		return wechatPaymentResource{}, fmt.Errorf("decode wechat resource: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return wechatPaymentResource{}, fmt.Errorf("create wechat cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return wechatPaymentResource{}, fmt.Errorf("create wechat gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, []byte(resource.Nonce), ciphertext, []byte(resource.AssociatedData))
	if err != nil {
		return wechatPaymentResource{}, fmt.Errorf("decrypt wechat callback: %w", ErrInvalidSignature)
	}
	var result wechatPaymentResource
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return wechatPaymentResource{}, fmt.Errorf("decode wechat resource: %w", err)
	}
	return result, nil
}

func (p *WeChatProvider) doSigned(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	privateKey, err := parseRSAPrivateKey(p.config.PrivateKey)
	if err != nil {
		return nil, err
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("generate wechat nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	bodyText := string(body)
	message := method + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + bodyText + "\n"
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return nil, fmt.Errorf("sign wechat request: %w", err)
	}
	authorization := fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`, p.config.MchID, nonce, timestamp, p.config.PublicKeyID, base64.StdEncoding.EncodeToString(signature))
	request, err := http.NewRequestWithContext(ctx, method, "https://api.mch.weixin.qq.com"+path, strings.NewReader(bodyText))
	if err != nil {
		return nil, fmt.Errorf("create wechat request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", authorization)
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("wechat request: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read wechat response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("wechat request returned status %d", response.StatusCode)
	}
	return responseBody, nil
}
