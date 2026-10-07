package payment

import (
	"errors"
	"net"
	"net/http"
	"strings"

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
	// Epay gateways notify via POST form or GET query depending on the
	// implementation, so both methods hit the same handlers.
	r.Post("/api/payment/callback/wechat", handler.wechatCallback)
	r.Get("/api/payment/callback/wechat", handler.wechatCallback)
	r.Post("/api/payment/callback/alipay", handler.alipayCallback)
	r.Get("/api/payment/callback/alipay", handler.alipayCallback)
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

// clientIP returns the payer IP for the epay clientip field: the first
// X-Forwarded-For hop when a proxy supplied one, otherwise the remote address
// without its port.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
			return first
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// methods handles GET /api/payment/methods.
//
// @Summary GET /api/payment/methods
// @ID payment_methods_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/methods [get]
func (h *Handler) methods(w http.ResponseWriter, r *http.Request) {
	methods, err := h.service.AvailableMethods(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, methods)
}

// createSubscription handles POST /api/payment/methods.
//
// @Summary POST /api/payment/methods
// @ID payment_methods_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/methods [post]
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
	result, err := h.service.CreateSubscriptionPaymentByMethod(r.Context(), principal.UserID, request.PlanID, request.Method, clientIP(r))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, result)
}

// recharge handles POST /api/recharge.
//
// @Summary POST /api/recharge
// @ID recharge_route_post
// @Tags recharge
// @Accept json
// @Produce json
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/recharge [post]
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
	result, err := h.service.CreateRechargePaymentByMethod(r.Context(), principal.UserID, request.Amount, request.Credits, request.Method, clientIP(r))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, result)
}

// history handles GET /api/payment/user.
//
// @Summary GET /api/payment/user
// @ID payment_user_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/user [get]
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

// settings handles GET /api/payment/settings.
//
// @Summary GET /api/payment/settings
// @ID payment_settings_route_get
// @Tags payment
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/settings [get]
func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.Settings(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, settings)
}

// get handles GET /api/payment/{outTradeNo}.
//
// @Summary GET /api/payment/{outTradeNo}
// @ID payment_outtradeno_route_get
// @Tags payment
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/{outTradeNo} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	h.respondOwnedPayment(w, r, func(outTradeNo string) (servicepayment.PaymentView, error) {
		return h.service.GetPayment(r.Context(), outTradeNo)
	})
}

// status handles GET /api/payment/{outTradeNo}/status.
//
// @Summary GET /api/payment/{outTradeNo}/status
// @ID payment_outtradeno_status_route_get
// @Tags payment
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/{outTradeNo}/status [get]
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

// cancel handles POST /api/payment/{outTradeNo}/cancel.
//
// @Summary POST /api/payment/{outTradeNo}/cancel
// @ID payment_outtradeno_cancel_route_post
// @Tags payment
// @Accept json
// @Produce json
// @Param outTradeNo path string true "outTradeNo"
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/payment/{outTradeNo}/cancel [post]
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
