package accounts

import (
	apierror "github.com/shuTwT/nex-api/backend/internal/service/apierror"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
)

var (
	ErrInvalidRequest = apierror.ErrValidation
	ErrNotFound       = apierror.ErrNotFound
	ErrConflict       = apierror.ErrConflict
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

type UserCreateRequest = model.UserCreateReq
type UserUpdateRequest = model.UserUpdateReq
type SubscriptionView = model.SubscriptionDTO
type UsageView = model.UsageDTO
type APIView = model.UsageAPIDTO
type UserView = model.UserResp
type UserStats = model.UserStatsResp

type UserListFilter struct {
	Role   string
	Search string
}
