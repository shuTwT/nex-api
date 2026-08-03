package membership

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	"github.com/shuTwT/nex-api/internal/middleware"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicemembership "github.com/shuTwT/nex-api/internal/service/membership"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type Handler struct {
	plans      *servicemembership.PlanService
	membership *servicemembership.MembershipService
	redemption *servicemembership.RedemptionService
}

func NewHandler(plans *servicemembership.PlanService, membership *servicemembership.MembershipService, redemption *servicemembership.RedemptionService) (*Handler, error) {
	if plans == nil || membership == nil || redemption == nil {
		return nil, errors.New("membership: all services are required")
	}
	return &Handler{plans: plans, membership: membership, redemption: redemption}, nil
}

func RegisterRoutes(r chi.Router, handler *Handler) error {
	if r == nil {
		return errors.New("membership: route mux is nil")
	}
	if handler == nil {
		return errors.New("membership: handler is nil")
	}
	h := handler
	admin := func(next http.Handler) http.Handler { return middleware.RequireAdmin(next) }
	user := func(next http.Handler) http.Handler { return middleware.RequireUser(next) }
	r.Method(http.MethodGet, "/api/subscription-plans", admin(http.HandlerFunc(h.listPlans)))
	r.Method(http.MethodPost, "/api/subscription-plans", admin(http.HandlerFunc(h.createPlan)))
	r.Method(http.MethodGet, "/api/subscription-plans/{id}", admin(http.HandlerFunc(h.getPlan)))
	r.Method(http.MethodPut, "/api/subscription-plans/{id}", admin(http.HandlerFunc(h.updatePlan)))
	r.Method(http.MethodDelete, "/api/subscription-plans/{id}", admin(http.HandlerFunc(h.deletePlan)))
	r.Get("/api/membership/plans", h.membershipPlans)
	r.Method(http.MethodGet, "/api/membership/current", user(http.HandlerFunc(h.currentMembership)))
	r.Method(http.MethodPost, "/api/membership/subscribe", user(http.HandlerFunc(h.subscribe)))
	r.Method(http.MethodGet, "/api/redemption-codes", admin(http.HandlerFunc(h.listCodes)))
	r.Method(http.MethodPost, "/api/redemption-codes", admin(http.HandlerFunc(h.createCodes)))
	r.Method(http.MethodDelete, "/api/redemption-codes/{id}", admin(http.HandlerFunc(h.deleteCode)))
	r.Method(http.MethodDelete, "/api/redemption-codes/batch", admin(http.HandlerFunc(h.deleteBatch)))
	r.Method(http.MethodPost, "/api/redemption-codes/batch", admin(http.HandlerFunc(h.deleteSelected)))
	r.Method(http.MethodGet, "/api/redemption-codes/export", admin(http.HandlerFunc(h.exportCodes)))
	r.Get("/api/redemption-codes/plans", h.redemptionPlans)
	r.Method(http.MethodPost, "/api/personal/redeem/lookup", user(http.HandlerFunc(h.lookupCode)))
	r.Method(http.MethodPost, "/api/personal/redeem", user(http.HandlerFunc(h.redeemCode)))
	return nil
}

func RegisterServiceRoutes(r chi.Router, plans *servicemembership.PlanService, membership *servicemembership.MembershipService, redemption *servicemembership.RedemptionService) error {
	handler, err := NewHandler(plans, membership, redemption)
	if err != nil {
		return err
	}
	return RegisterRoutes(r, handler)
}

func (h *Handler) membershipPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.membership.ListPlans(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, plans)
}

func (h *Handler) currentMembership(w http.ResponseWriter, r *http.Request) {
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	current, err := h.membership.Current(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, current)
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.MembershipSubscribeReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	subscription, err := h.membership.Subscribe(r.Context(), principal.UserID, body.PlanID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, subscription)
}

func planFilter(r *http.Request) (servicemembership.PlanListFilter, error) {
	page, limit, err := pageQuery(r)
	if err != nil {
		return servicemembership.PlanListFilter{}, err
	}
	var active *bool
	if raw := r.URL.Query().Get("isActive"); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			return servicemembership.PlanListFilter{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "isActive", Reason: "must be boolean"})
		}
		active = &value
	}
	return servicemembership.PlanListFilter{Search: r.URL.Query().Get("search"), IsActive: active, Page: page, Limit: limit}, nil
}

func redemptionFilter(r *http.Request) (servicemembership.RedemptionListFilter, error) {
	page, limit, err := pageQuery(r)
	if err != nil {
		return servicemembership.RedemptionListFilter{}, err
	}
	var used *bool
	if raw := r.URL.Query().Get("isUsed"); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			return servicemembership.RedemptionListFilter{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "isUsed", Reason: "must be boolean"})
		}
		used = &value
	}
	return servicemembership.RedemptionListFilter{Search: r.URL.Query().Get("search"), Type: r.URL.Query().Get("type"), IsUsed: used, Page: page, Limit: limit}, nil
}

func pageQuery(r *http.Request) (int, int, error) {
	page, limit := 1, 10
	var err error
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			return 0, 0, appRuntime.NewValidationError(appRuntime.FieldError{Field: "page", Reason: "must be positive"})
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, appRuntime.NewValidationError(appRuntime.FieldError{Field: "limit", Reason: "must be between 1 and 100"})
		}
	}
	return page, limit, nil
}
