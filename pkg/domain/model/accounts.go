// Package model contains the API's shared request, response, and DTO models.
package model

import "time"

type UserCreateReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Credits  int    `json:"credits"`
}
type UserUpdateReq struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Role     *string `json:"role"`
	Credits  *int    `json:"credits"`
}
type SubscriptionDTO struct {
	ID        string    `json:"id"`
	PlanName  string    `json:"planName"`
	Credits   int       `json:"credits"`
	Price     float64   `json:"price"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	IsActive  bool      `json:"isActive"`
}
type UsageAPIDTO struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}
type UsageDTO struct {
	ID        string       `json:"id"`
	Credits   int          `json:"credits"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"createdAt"`
	API       *UsageAPIDTO `json:"api,omitempty"`
}
type UserResp struct {
	ID           string           `json:"id"`
	Name         string           `json:"name,omitempty"`
	Email        string           `json:"email"`
	Username     string           `json:"username"`
	Role         string           `json:"role"`
	Credits      int              `json:"credits"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	Subscription *SubscriptionDTO `json:"subscription,omitempty"`
	APIUsage     []UsageDTO       `json:"apiUsage,omitempty"`
}
type UserStatsResp struct {
	TotalUsers        int `json:"totalUsers"`
	ActiveUsers       int `json:"activeUsers"`
	AdminUsers        int `json:"adminUsers"`
	NewUsersThisMonth int `json:"newUsersThisMonth"`
}

type TokenCreateReq struct {
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"`
	ExpiresAt   *time.Time `json:"expiresAt"`
}
type TokenUpdateReq struct {
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	IsActive    *bool      `json:"isActive"`
}
type TokenResp struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Permissions string     `json:"permissions"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
type TokenCreateResp struct {
	TokenResp
	Token string `json:"token"`
}
type TokenStatsResp struct {
	TotalTokens    int `json:"totalTokens"`
	ActiveTokens   int `json:"activeTokens"`
	InactiveTokens int `json:"inactiveTokens"`
	ExpiredTokens  int `json:"expiredTokens"`
}

type ProfileUpdateReq struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
}
type PasswordUpdateReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}
type ProfileResp struct {
	ID                string    `json:"id"`
	Name              string    `json:"name,omitempty"`
	Email             string    `json:"email"`
	Image             string    `json:"image,omitempty"`
	Username          string    `json:"username"`
	Role              string    `json:"role"`
	Credits           int       `json:"credits"`
	CreatedAt         time.Time `json:"createdAt"`
	TotalCreditsSpent int       `json:"totalCreditsSpent"`
	TotalRequests     int       `json:"totalRequests"`
}

type AuditUserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type AuditResp struct {
	ID        string        `json:"id"`
	UserID    string        `json:"userId,omitempty"`
	User      *AuditUserDTO `json:"user,omitempty"`
	Action    string        `json:"action"`
	Resource  string        `json:"resource"`
	Details   string        `json:"details,omitempty"`
	IPAddress string        `json:"ipAddress,omitempty"`
	UserAgent string        `json:"userAgent,omitempty"`
	Level     string        `json:"level"`
	Status    string        `json:"status"`
	Metadata  string        `json:"metadata,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
}
type AuditStatsResp struct {
	TotalLogs   int `json:"totalLogs"`
	InfoLogs    int `json:"infoLogs"`
	WarningLogs int `json:"warningLogs"`
	ErrorLogs   int `json:"errorLogs"`
	SuccessLogs int `json:"successLogs"`
	FailedLogs  int `json:"failedLogs"`
}
