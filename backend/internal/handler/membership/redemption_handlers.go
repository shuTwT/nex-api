package membership

import (
	servicemembership "github.com/shuTwT/nex-api/backend/internal/service/membership"
	"net/http"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/service/authz"
)

func (h *Handler) listCodes(w http.ResponseWriter, r *http.Request) {
	filter, err := redemptionFilter(r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	page, err := h.redemption.List(r.Context(), filter)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writePage(w, http.StatusOK, page.Items, page.Total, page.Page, page.PageSize, page.TotalPages)
}

func (h *Handler) createCodes(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[redemptionCreateBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
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
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, result)
}

func (h *Handler) deleteCode(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	if err := h.redemption.Delete(r.Context(), principal.UserID, r.PathValue("id")); err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "兑换码已删除"})
}

func (h *Handler) deleteBatch(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	count, err := h.redemption.DeleteBatch(r.Context(), principal.UserID, r.URL.Query().Get("batchId"))
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) deleteSelected(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[idsBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	count, err := h.redemption.DeleteBatchByIDs(r.Context(), principal.UserID, body.IDs)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) exportCodes(w http.ResponseWriter, r *http.Request) {
	content, err := h.redemption.Export(r.Context(), strings.Split(r.URL.Query().Get("ids"), ","))
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, content)
}

func (h *Handler) redemptionPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.redemption.ListPlans(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, plans)
}

func (h *Handler) lookupCode(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[codeBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	lookup, err := h.redemption.Lookup(r.Context(), body.Code)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, lookup)
}

func (h *Handler) redeemCode(w http.ResponseWriter, r *http.Request) {
	body, err := decodeJSON[codeBody](r)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	result, err := h.redemption.Redeem(r.Context(), principal.UserID, body.Code)
	if err != nil {
		writeMembershipError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

type redemptionCreateBody struct {
	Type      string     `json:"type"`
	Count     *int       `json:"count"`
	PlanID    string     `json:"planId"`
	Credits   *int       `json:"credits"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
