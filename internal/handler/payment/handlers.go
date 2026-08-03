package payment

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/internal/middleware"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicepayment "github.com/shuTwT/nex-api/internal/service/payment"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type Handler struct {
	service *servicepayment.Service
}

func NewHandler(service *servicepayment.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("payment: service is nil")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(r chi.Router, handler *Handler) error {
	if r == nil || handler == nil {
		return errors.New("payment: mux and handler are required")
	}
	user := func(next http.Handler) http.Handler { return middleware.RequireUser(next) }
	r.Get("/api/payment/methods", handler.methods)
	r.Method(http.MethodPost, "/api/payment/methods", user(http.HandlerFunc(handler.createSubscription)))
	r.Method(http.MethodGet, "/api/payment/user", user(http.HandlerFunc(handler.history)))
	r.Method(http.MethodGet, "/api/payment/settings", user(http.HandlerFunc(handler.settings)))
	r.Method(http.MethodGet, "/api/payment/{outTradeNo}/status", user(http.HandlerFunc(handler.status)))
	r.Method(http.MethodGet, "/api/payment/{outTradeNo}", user(http.HandlerFunc(handler.get)))
	r.Method(http.MethodPost, "/api/payment/{outTradeNo}/cancel", user(http.HandlerFunc(handler.cancel)))
	r.Method(http.MethodPost, "/api/recharge", user(http.HandlerFunc(handler.recharge)))
	r.Post("/api/payment/callback/wechat", handler.wechatCallback)
	r.Post("/api/payment/callback/alipay", handler.alipayCallback)
	r.Post("/api/payment/callback/mock", handler.mockCallback)
	return nil
}

func RegisterServiceRoutes(r chi.Router, service *servicepayment.Service) error {
	handler, err := NewHandler(service)
	if err != nil {
		return err
	}
	return RegisterRoutes(r, handler)
}

func (h *Handler) methods(w http.ResponseWriter, r *http.Request) {
	methods, err := h.service.AvailableMethods(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, methods)
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	request, err := handlerutils.DecodeJSONValue[model.SubscriptionPaymentCreateReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.service.CreateSubscriptionPaymentByMethod(r.Context(), principal.UserID, request.PlanID, request.Method)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, result)
}

func (h *Handler) recharge(w http.ResponseWriter, r *http.Request) {
	request, err := handlerutils.DecodeJSONValue[model.RechargePaymentCreateReq](r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.service.CreateRechargePaymentByMethod(r.Context(), principal.UserID, request.Amount, request.Credits, request.Method)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, result)
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	payments, err := h.service.ListHistory(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, payments)
}

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.Settings(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, settings)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	h.respondOwnedPayment(w, r, func(outTradeNo string) (servicepayment.PaymentView, error) {
		return h.service.GetPayment(r.Context(), outTradeNo)
	})
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	h.respondOwnedPayment(w, r, func(outTradeNo string) (servicepayment.PaymentView, error) {
		return h.service.QueryPayment(r.Context(), outTradeNo)
	})
}

func (h *Handler) respondOwnedPayment(w http.ResponseWriter, r *http.Request, load func(string) (servicepayment.PaymentView, error)) {
	outTradeNo := r.PathValue("outTradeNo")
	view, err := load(outTradeNo)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := serviceauthz.CheckOwnership(principal, view.UserID); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	outTradeNo := r.PathValue("outTradeNo")
	view, err := h.service.GetPayment(r.Context(), outTradeNo)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := serviceauthz.CheckOwnership(principal, view.UserID); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := h.service.CancelPayment(r.Context(), outTradeNo); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "payment cancelled"})
}

func writeRawJSON(w http.ResponseWriter, status int, value string) {
	writeRaw(w, status, "application/json", value)
}

func writeRaw(w http.ResponseWriter, status int, contentType, value string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(value))
}
