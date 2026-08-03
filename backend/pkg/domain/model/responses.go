package model

import (
	"time"

	pay "github.com/shuTwT/nex-api/backend/internal/infra/pay"
)

type PaymentResp struct {
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
type PaymentCreateResp struct {
	Success    bool   `json:"success"`
	PaymentID  string `json:"paymentId,omitempty"`
	OutTradeNo string `json:"outTradeNo,omitempty"`
	QRCodeURL  string `json:"qrcodeUrl,omitempty"`
	PayURL     string `json:"payUrl,omitempty"`
}
type PaymentSettingsResp struct {
	CreditPrice   float64 `json:"creditPrice"`
	MinRecharge   float64 `json:"minRecharge"`
	AlipayEnabled bool    `json:"alipayEnabled"`
	WeChatEnabled bool    `json:"wechatEnabled"`
	MockEnabled   bool    `json:"mockEnabled"`
}
type AdvertisementPositionDTO struct {
	Position string         `json:"position"`
	Count    map[string]int `json:"_count"`
}
type AdvertisementStatsResp struct {
	TotalAds      int                        `json:"totalAds"`
	ActiveAds     int                        `json:"activeAds"`
	InactiveAds   int                        `json:"inactiveAds"`
	PositionStats []AdvertisementPositionDTO `json:"positionStats"`
}
type RedemptionBatchResp struct {
	Count   int    `json:"count"`
	BatchID string `json:"batchId"`
}
type RedemptionLookupResp struct {
	Type     string  `json:"type"`
	PlanName *string `json:"planName"`
	Credits  *int    `json:"credits"`
}
type RedemptionResp struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Credits int    `json:"credits,omitempty"`
}
type DashboardActivityResp struct {
	ID        string `json:"id"`
	APIName   string `json:"apiName"`
	APIAlias  string `json:"apiAlias"`
	Credits   int    `json:"credits"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}
type DashboardTopAPIResp struct {
	Name       string `json:"name"`
	Calls      int64  `json:"calls"`
	Percentage int    `json:"percentage"`
}
