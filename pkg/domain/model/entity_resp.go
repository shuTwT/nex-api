package model

import "time"

// The Resp DTOs in this file mirror ent entities that handlers return
// directly. Swaggo cannot resolve ent package types from handler annotations,
// so the spec references these DTOs instead; dto_compat_test.go keeps their
// JSON shape locked to the entities they document.

type AdvertisementResp struct {
	ID          string    `json:"id,omitempty"`
	Image       string    `json:"image,omitempty"`
	ImageWidth  int       `json:"imageWidth,omitempty"`
	ImageHeight int       `json:"imageHeight,omitempty"`
	Link        string    `json:"link,omitempty"`
	Title       string    `json:"title,omitempty"`
	Position    string    `json:"position,omitempty"`
	IsActive    bool      `json:"isActive,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

type SubscriptionPlanResp struct {
	ID               string    `json:"id,omitempty"`
	Title            string    `json:"title,omitempty"`
	Price            float64   `json:"price"`
	TotalCredits     int       `json:"totalCredits,omitempty"`
	SortOrder        int       `json:"sortOrder,omitempty"`
	ValidityDuration int       `json:"validityDuration,omitempty"`
	ValidityUnit     string    `json:"validityUnit,omitempty"`
	CreditResetCycle string    `json:"creditResetCycle,omitempty"`
	IsActive         bool      `json:"isActive,omitempty"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt,omitempty"`
}

type SubscriptionResp struct {
	ID        string    `json:"id,omitempty"`
	UserId    string    `json:"userId,omitempty"`
	PlanId    string    `json:"planId,omitempty"`
	PlanName  string    `json:"planName,omitempty"`
	Credits   int       `json:"credits,omitempty"`
	Price     float64   `json:"price,omitempty"`
	StartDate time.Time `json:"startDate,omitempty"`
	EndDate   time.Time `json:"endDate,omitempty"`
	IsActive  bool      `json:"isActive,omitempty"`
	PaymentId string    `json:"paymentId,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type RedemptionCodeResp struct {
	ID        string    `json:"id,omitempty"`
	Code      string    `json:"code,omitempty"`
	Type      string    `json:"type,omitempty"`
	PlanId    string    `json:"planId,omitempty"`
	PlanName  string    `json:"planName,omitempty"`
	Credits   int       `json:"credits,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	IsUsed    bool      `json:"isUsed,omitempty"`
	UsedBy    string    `json:"usedBy,omitempty"`
	UsedAt    time.Time `json:"usedAt,omitempty"`
	CreatedBy string    `json:"createdBy,omitempty"`
	BatchId   string    `json:"batchId,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type SystemSettingResp struct {
	ID          string    `json:"id,omitempty"`
	Key         string    `json:"key,omitempty"`
	Value       string    `json:"value,omitempty"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}
