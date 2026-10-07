package membership

import (
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicemembership "github.com/shuTwT/nex-api/internal/service/membership"
	"github.com/shuTwT/nex-api/pkg/domain/model"
	"net/http"
	"strings"

	"github.com/shuTwT/nex-api/internal/service/authz"
)

// listCodes handles GET /api/redemption-codes.
//
// @Summary GET /api/redemption-codes
// @ID redemption_codes_route_get
// @Tags redemption-codes
// @Produce json
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.RedemptionCodeResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes [get]
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

// createCodes handles POST /api/redemption-codes.
//
// @Summary POST /api/redemption-codes
// @ID redemption_codes_route_post
// @Tags redemption-codes
// @Accept json
// @Produce json
// @Param body body model.RedemptionCodeCreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.RedemptionBatchResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes [post]
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

// deleteCode handles DELETE /api/redemption-codes/{id}.
//
// @Summary DELETE /api/redemption-codes/{id}
// @ID redemption_codes_id_route_delete
// @Tags redemption-codes
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes/{id} [delete]
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

// deleteBatch handles DELETE /api/redemption-codes/batch.
//
// @Summary DELETE /api/redemption-codes/batch
// @ID redemption_codes_batch_route_delete
// @Tags redemption-codes
// @Produce json
// @Param batchId query string false "batchId"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]int}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes/batch [delete]
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

// deleteSelected handles POST /api/redemption-codes/batch.
//
// @Summary POST /api/redemption-codes/batch
// @ID redemption_codes_batch_route_post
// @Tags redemption-codes
// @Accept json
// @Produce json
// @Param body body model.IDsReq false "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes/batch [post]
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

// exportCodes handles GET /api/redemption-codes/export.
//
// @Summary GET /api/redemption-codes/export
// @ID redemption_codes_export_route_get
// @Tags redemption-codes
// @Produce json
// @Param ids query string false "ids"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes/export [get]
func (h *Handler) exportCodes(w http.ResponseWriter, r *http.Request) {
	content, err := h.redemption.Export(r.Context(), strings.Split(r.URL.Query().Get("ids"), ","))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, content)
}

// redemptionPlans handles GET /api/redemption-codes/plans.
//
// @Summary GET /api/redemption-codes/plans
// @ID redemption_codes_plans_route_get
// @Tags redemption-codes
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/redemption-codes/plans [get]
func (h *Handler) redemptionPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.redemption.ListPlans(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, plans)
}

// lookupCode handles POST /api/personal/redeem/lookup.
//
// @Summary POST /api/personal/redeem/lookup
// @ID personal_redeem_lookup_route_post
// @Tags personal
// @Accept json
// @Produce json
// @Param body body model.RedemptionCodeReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.RedemptionLookupResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/personal/redeem/lookup [post]
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

// redeemCode handles POST /api/personal/redeem.
//
// @Summary POST /api/personal/redeem
// @ID personal_redeem_route_post
// @Tags personal
// @Accept json
// @Produce json
// @Param body body model.RedemptionCodeReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.RedemptionResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/personal/redeem [post]
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
