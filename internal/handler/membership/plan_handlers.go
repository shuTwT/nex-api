package membership

import (
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicemembership "github.com/shuTwT/nex-api/internal/service/membership"
	"github.com/shuTwT/nex-api/pkg/domain/model"
	"net/http"

	"github.com/shuTwT/nex-api/internal/service/authz"
)

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	filter, err := planFilter(r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	page, err := h.plans.List(r.Context(), filter)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, page.Items, page.Page, page.PageSize, page.Total)
}

func (h *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := h.plans.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, plan)
}

func (h *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.SubscriptionPlanCreateReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
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
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, plan)
}

func (h *Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	input, err := handlerutils.DecodeJSONValue[servicemembership.PlanUpdateInput](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	plan, err := h.plans.Update(r.Context(), principal.UserID, r.PathValue("id"), input)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, plan)
}

func (h *Handler) deletePlan(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := h.plans.Delete(r.Context(), principal.UserID, r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "订阅计划已删除"})
}
