package membership

import (
	servicemembership "github.com/shuTwT/nex-api/backend/internal/service/membership"
	"net/http"

	"github.com/shuTwT/nex-api/backend/internal/service/authz"
)

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	filter, err := planFilter(r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	page, err := h.plans.List(r.Context(), filter)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writePage(w, http.StatusOK, page.Items, page.Total, page.Page, page.PageSize, page.TotalPages)
}

func (h *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := h.plans.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, plan)
}

func (h *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[planCreateBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	input := servicemembership.PlanCreateInput{Title: body.Title, IsActive: false}
	if body.Price != nil {
		input.Price = *body.Price
	}
	if body.TotalCredits != nil {
		input.TotalCredits = *body.TotalCredits
	}
	if body.SortOrder != nil {
		input.SortOrder = *body.SortOrder
	}
	input.ValidityDuration = 30
	if body.ValidityDuration != nil {
		input.ValidityDuration = *body.ValidityDuration
	}
	input.ValidityUnit = body.ValidityUnit
	input.CreditResetCycle = body.CreditResetCycle
	if body.IsActive != nil {
		input.IsActive = *body.IsActive
	}
	plan, err := h.plans.Create(r.Context(), principal.UserID, input)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, plan)
}

func (h *Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	input, err := decodeJSON[servicemembership.PlanUpdateInput](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	plan, err := h.plans.Update(r.Context(), principal.UserID, r.PathValue("id"), input)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, plan)
}

func (h *Handler) deletePlan(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	if err := h.plans.Delete(r.Context(), principal.UserID, r.PathValue("id")); err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "订阅计划已删除"})
}

type planCreateBody struct {
	Title            string   `json:"title"`
	Price            *float64 `json:"price"`
	TotalCredits     *int     `json:"totalCredits"`
	SortOrder        *int     `json:"sortOrder"`
	ValidityDuration *int     `json:"validityDuration"`
	ValidityUnit     string   `json:"validityUnit"`
	CreditResetCycle string   `json:"creditResetCycle"`
	IsActive         *bool    `json:"isActive"`
}
