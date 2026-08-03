package payment

import (
	"encoding/json"
	"time"

	pay "github.com/shuTwT/nex-api/backend/internal/infra/pay"
)

type createPaymentInput struct {
	UserID    string
	Amount    float64
	Currency  string
	Method    pay.PaymentMethod
	PlanID    string
	NotifyURL string
	Metadata  map[string]json.RawMessage
}

type PaymentView struct {
	ID            string            `json:"id"`
	UserID        string            `json:"userId"`
	OutTradeNo    string            `json:"outTradeNo"`
	TransactionID string            `json:"transactionId,omitempty"`
	Method        pay.PaymentMethod `json:"method"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Status        pay.PaymentState  `json:"status"`
	QRCodeURL     string            `json:"qrcodeUrl,omitempty"`
	PayURL        string            `json:"payUrl,omitempty"`
	PaidAt        *time.Time        `json:"paidAt,omitempty"`
	ExpiredAt     *time.Time        `json:"expiredAt,omitempty"`
	CancelledAt   *time.Time        `json:"cancelledAt,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type CreatePaymentResult struct {
	Success    bool   `json:"success"`
	PaymentID  string `json:"paymentId,omitempty"`
	OutTradeNo string `json:"outTradeNo,omitempty"`
	QRCodeURL  string `json:"qrcodeUrl,omitempty"`
	PayURL     string `json:"payUrl,omitempty"`
}

type PaymentSettings struct {
	CreditPrice   float64 `json:"creditPrice"`
	MinRecharge   float64 `json:"minRecharge"`
	AlipayEnabled bool    `json:"alipayEnabled"`
	WeChatEnabled bool    `json:"wechatEnabled"`
	MockEnabled   bool    `json:"mockEnabled"`
}
