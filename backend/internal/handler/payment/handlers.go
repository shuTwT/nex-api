package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicepayment "github.com/shuTwT/nex-api/backend/internal/service/payment"
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

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("payment: mux and handler are required")
	}
	user := func(next http.Handler) http.Handler { return middleware.RequireUser(next) }
	mux.HandleFunc("GET /api/payment/methods", handler.methods)
	mux.Handle("POST /api/payment/methods", user(http.HandlerFunc(handler.createSubscription)))
	mux.Handle("GET /api/payment/user", user(http.HandlerFunc(handler.history)))
	mux.Handle("GET /api/payment/settings", user(http.HandlerFunc(handler.settings)))
	mux.Handle("GET /api/payment/{outTradeNo}/status", user(http.HandlerFunc(handler.status)))
	mux.Handle("GET /api/payment/{outTradeNo}", user(http.HandlerFunc(handler.get)))
	mux.Handle("POST /api/payment/{outTradeNo}/cancel", user(http.HandlerFunc(handler.cancel)))
	mux.Handle("POST /api/recharge", user(http.HandlerFunc(handler.recharge)))
	mux.HandleFunc("POST /api/payment/callback/wechat", handler.wechatCallback)
	mux.HandleFunc("POST /api/payment/callback/alipay", handler.alipayCallback)
	mux.HandleFunc("POST /api/payment/callback/mock", handler.mockCallback)
	return nil
}

func RegisterServiceRoutes(mux *http.ServeMux, service *servicepayment.Service) error {
	handler, err := NewHandler(service)
	if err != nil {
		return err
	}
	return RegisterRoutes(mux, handler)
}

type subscriptionPaymentRequest struct {
	PlanID string `json:"planId"`
	Method string `json:"method"`
}

type rechargeRequest struct {
	Amount  float64 `json:"amount"`
	Credits int     `json:"credits"`
	Method  string  `json:"method"`
}

func (h *Handler) methods(w http.ResponseWriter, r *http.Request) {
	methods, err := h.service.AvailableMethods(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, methods)
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	request, err := decodeJSON[subscriptionPaymentRequest](r)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	result, err := h.service.CreateSubscriptionPaymentByMethod(r.Context(), principal.UserID, request.PlanID, request.Method)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, result)
}

func (h *Handler) recharge(w http.ResponseWriter, r *http.Request) {
	request, err := decodeJSON[rechargeRequest](r)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	result, err := h.service.CreateRechargePaymentByMethod(r.Context(), principal.UserID, request.Amount, request.Credits, request.Method)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, result)
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	payments, err := h.service.ListHistory(r.Context(), principal.UserID)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, payments)
}

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.Settings(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, settings)
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
		writePaymentError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	if err := serviceauthz.CheckOwnership(principal, view.UserID); err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	outTradeNo := r.PathValue("outTradeNo")
	view, err := h.service.GetPayment(r.Context(), outTradeNo)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := serviceauthz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	if err := serviceauthz.CheckOwnership(principal, view.UserID); err != nil {
		writePaymentError(w, r, err)
		return
	}
	if err := h.service.CancelPayment(r.Context(), outTradeNo); err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "payment cancelled"})
}

func decodeJSON[T interface{}](r *http.Request) (T, error) {
	var value T
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode payment request: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return value, errors.New("decode payment request: multiple JSON values")
	}
	return value, nil
}

func writeData[T interface{}](w http.ResponseWriter, status int, value T) {
	_ = appRuntime.WriteData(w, status, value)
}

func writePaymentError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}

func writeRawJSON(w http.ResponseWriter, status int, value string) {
	writeRaw(w, status, "application/json", value)
}

func writeRaw(w http.ResponseWriter, status int, contentType, value string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(value))
}
