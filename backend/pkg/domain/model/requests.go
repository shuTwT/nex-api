package model

import "time"

type AuthLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AdvertisementCreateReq struct {
	Image       string `json:"image"`
	ImageWidth  int    `json:"imageWidth"`
	ImageHeight int    `json:"imageHeight"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Position    string `json:"position"`
	IsActive    *bool  `json:"isActive"`
}
type AdvertisementUpdateReq struct {
	Image       *string `json:"image"`
	ImageWidth  *int    `json:"imageWidth"`
	ImageHeight *int    `json:"imageHeight"`
	Link        *string `json:"link"`
	Title       *string `json:"title"`
	Position    *string `json:"position"`
	IsActive    *bool   `json:"isActive"`
}
type SubscriptionPaymentCreateReq struct {
	PlanID string `json:"planId"`
	Method string `json:"method"`
}
type RechargePaymentCreateReq struct {
	Amount  float64 `json:"amount"`
	Credits int     `json:"credits"`
	Method  string  `json:"method"`
}
type MembershipSubscribeReq struct {
	PlanID string `json:"planId"`
}
type RedemptionCodeReq struct {
	Code string `json:"code"`
}
type IDsReq struct {
	IDs []string `json:"ids"`
}
type RedemptionCodeCreateReq struct {
	Type      string     `json:"type"`
	Count     *int       `json:"count"`
	PlanID    string     `json:"planId"`
	Credits   *int       `json:"credits"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
type SubscriptionPlanCreateReq struct {
	Title            string   `json:"title"`
	Price            *float64 `json:"price"`
	TotalCredits     *int     `json:"totalCredits"`
	SortOrder        *int     `json:"sortOrder"`
	ValidityDuration *int     `json:"validityDuration"`
	ValidityUnit     string   `json:"validityUnit"`
	CreditResetCycle string   `json:"creditResetCycle"`
	IsActive         *bool    `json:"isActive"`
}
