package membership

import (
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicemembership "github.com/shuTwT/nex-api/internal/service/membership"
	"github.com/shuTwT/nex-api/pkg/domain/model"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/internal/service/authz"
)

func (h *Handler) listCodes(w http.ResponseWriter, r *http.Request) {
	filter, err := redemptionFilter(r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	page, err := h.redemption.List(r.Context(), filter)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, page.Items, page.Page, page.PageSize, page.Total)
}

func (h *Handler) createCodes(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.RedemptionCodeCreateReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	count := 1
	if body.Count != nil {
		count = *body.Count
	}
	credits := 0
	if body.Credits != nil {
		credits = *body.Credits
	}
	result, err := h.redemption.CreateBatch(r.Context(), principal.UserID, servicemembership.RedemptionCreateInput{Type: body.Type, Count: count, PlanID: body.PlanID, Credits: credits, ExpiresAt: body.ExpiresAt})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, result)
}

func (h *Handler) deleteCode(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := h.redemption.Delete(r.Context(), principal.UserID, r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "兑换码已删除"})
}

func (h *Handler) deleteBatch(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	count, err := h.redemption.DeleteBatch(r.Context(), principal.UserID, r.URL.Query().Get("batchId"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) deleteSelected(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.IDsReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	count, err := h.redemption.DeleteBatchByIDs(r.Context(), principal.UserID, body.IDs)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) exportCodes(w http.ResponseWriter, r *http.Request) {
	content, err := h.redemption.Export(r.Context(), strings.Split(r.URL.Query().Get("ids"), ","))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, content)
}

func (h *Handler) redemptionPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.redemption.ListPlans(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, plans)
}

func (h *Handler) lookupCode(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.RedemptionCodeReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	lookup, err := h.redemption.Lookup(r.Context(), body.Code)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, lookup)
}

func (h *Handler) redeemCode(w http.ResponseWriter, r *http.Request) {
	body, err := handlerutils.DecodeJSONValue[model.RedemptionCodeReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.redemption.Redeem(r.Context(), principal.UserID, body.Code)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, result)
}
