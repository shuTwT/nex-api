package membership

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicemembership "github.com/shuTwT/nex-api/backend/internal/service/membership"
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

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil {
		return errors.New("membership: route mux is nil")
	}
	if handler == nil {
		return errors.New("membership: handler is nil")
	}
	h := handler
	admin := func(next http.Handler) http.Handler { return middleware.RequireAdmin(next) }
	user := func(next http.Handler) http.Handler { return middleware.RequireUser(next) }
	mux.Handle("GET /api/subscription-plans", admin(http.HandlerFunc(h.listPlans)))
	mux.Handle("POST /api/subscription-plans", admin(http.HandlerFunc(h.createPlan)))
	mux.Handle("GET /api/subscription-plans/{id}", admin(http.HandlerFunc(h.getPlan)))
	mux.Handle("PUT /api/subscription-plans/{id}", admin(http.HandlerFunc(h.updatePlan)))
	mux.Handle("DELETE /api/subscription-plans/{id}", admin(http.HandlerFunc(h.deletePlan)))
	mux.HandleFunc("GET /api/membership/plans", h.membershipPlans)
	mux.Handle("GET /api/membership/current", user(http.HandlerFunc(h.currentMembership)))
	mux.Handle("POST /api/membership/subscribe", user(http.HandlerFunc(h.subscribe)))
	mux.Handle("GET /api/redemption-codes", admin(http.HandlerFunc(h.listCodes)))
	mux.Handle("POST /api/redemption-codes", admin(http.HandlerFunc(h.createCodes)))
	mux.Handle("DELETE /api/redemption-codes/{id}", admin(http.HandlerFunc(h.deleteCode)))
	mux.Handle("DELETE /api/redemption-codes/batch", admin(http.HandlerFunc(h.deleteBatch)))
	mux.Handle("POST /api/redemption-codes/batch", admin(http.HandlerFunc(h.deleteSelected)))
	mux.Handle("GET /api/redemption-codes/export", admin(http.HandlerFunc(h.exportCodes)))
	mux.HandleFunc("GET /api/redemption-codes/plans", h.redemptionPlans)
	mux.Handle("POST /api/personal/redeem/lookup", user(http.HandlerFunc(h.lookupCode)))
	mux.Handle("POST /api/personal/redeem", user(http.HandlerFunc(h.redeemCode)))
	return nil
}

func RegisterServiceRoutes(mux *http.ServeMux, plans *servicemembership.PlanService, membership *servicemembership.MembershipService, redemption *servicemembership.RedemptionService) error {
	handler, err := NewHandler(plans, membership, redemption)
	if err != nil {
		return err
	}
	return RegisterRoutes(mux, handler)
}

func (h *Handler) membershipPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.membership.ListPlans(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, plans)
}

func (h *Handler) currentMembership(w http.ResponseWriter, r *http.Request) {
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	current, err := h.membership.Current(r.Context(), principal.UserID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, current)
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[subscribeBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	subscription, err := h.membership.Subscribe(r.Context(), principal.UserID, body.PlanID)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, subscription)
}

type subscribeBody struct {
	PlanID string `json:"planId"`
}
type codeBody struct {
	Code string `json:"code"`
}
type idsBody struct {
	IDs []string `json:"ids"`
}

func decodeJSON[T interface{}](r *http.Request) (T, error) {
	var value T
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode request: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return value, errors.New("decode request: multiple JSON values")
	}
	return value, nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	if err := appRuntime.WriteData(w, status, data); err != nil {
		return
	}
}

func writePage[T any](w http.ResponseWriter, status int, items T, total, page, size, pages int) {
	envelope := appRuntime.Envelope[T]{Success: true, Data: &items, Pagination: &appRuntime.Pagination{Page: page, PageSize: size, Total: total, TotalPages: pages}}
	if err := appRuntime.WriteEnvelope(w, status, envelope); err != nil {
		return
	}
}

func writeMembershipError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = errors.New("membership: empty error")
	}
	if writeErr := appRuntime.WriteError(w, r, err); writeErr != nil {
		return
	}
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
