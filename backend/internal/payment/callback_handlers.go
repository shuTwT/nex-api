package payment

import (
	"io"
	"net/http"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

func (h *Handler) wechatCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	_, err = h.service.HandleCallback(r.Context(), PaymentMethodWeChat, ProviderCallbackRequest{Body: body, Headers: r.Header})
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	writeRawJSON(w, http.StatusOK, `{"code":"SUCCESS","message":"成功"}`)
}

func (h *Handler) alipayCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.callbackError(w, r, err)
		return
	}
	_, err := h.service.HandleCallback(r.Context(), PaymentMethodAlipay, ProviderCallbackRequest{Form: r.PostForm})
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	writeRaw(w, http.StatusOK, "text/plain; charset=utf-8", "success")
}

func (h *Handler) mockCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	_, err = h.service.HandleCallback(r.Context(), PaymentMethodMock, ProviderCallbackRequest{Body: body})
	if err != nil {
		h.callbackError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, struct {
		Success bool `json:"success"`
	}{Success: true})
}

func (h *Handler) callbackError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}
