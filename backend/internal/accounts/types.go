package accounts

import (
	"time"

	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

var (
	ErrInvalidRequest = runtime.ErrValidation
	ErrNotFound       = runtime.ErrNotFound
	ErrConflict       = runtime.ErrConflict
)

type PageRequest struct {
	Page int
	Size int
}

func (p PageRequest) normalized() PageRequest {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size < 1 || p.Size > 100 {
		p.Size = 10
	}
	return p
}

type PageInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func pageInfo(page PageRequest, total int) PageInfo {
	page = page.normalized()
	return PageInfo{Page: page.Page, PageSize: page.Size, Total: total, TotalPages: (total + page.Size - 1) / page.Size}
}

type AuditMetadata struct {
	IP        string
	UserAgent string
	Metadata  string
}

type UserCreateRequest struct {
	Email    string
	Username string
	Password string
	Role     string
	Credits  int
}

type UserUpdateRequest struct {
	Email    *string
	Username *string
	Role     *string
	Credits  *int
}

type SubscriptionView struct {
	ID        string    `json:"id"`
	PlanName  string    `json:"planName"`
	Credits   int       `json:"credits"`
	Price     float64   `json:"price"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	IsActive  bool      `json:"isActive"`
}

type UsageView struct {
	ID        string    `json:"id"`
	Credits   int       `json:"credits"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	API       *APIView  `json:"api,omitempty"`
}

type APIView struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type UserView struct {
	ID           string            `json:"id"`
	Name         string            `json:"name,omitempty"`
	Email        string            `json:"email"`
	Username     string            `json:"username"`
	Role         string            `json:"role"`
	Credits      int               `json:"credits"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	Subscription *SubscriptionView `json:"subscription,omitempty"`
	APIUsage     []UsageView       `json:"apiUsage,omitempty"`
}

type UserStats struct {
	TotalUsers        int `json:"totalUsers"`
	ActiveUsers       int `json:"activeUsers"`
	AdminUsers        int `json:"adminUsers"`
	NewUsersThisMonth int `json:"newUsersThisMonth"`
}

type UserListFilter struct {
	Role   string
	Search string
}
