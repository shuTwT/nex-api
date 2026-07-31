package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/shuTwT/nex-api/backend/internal/authz"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("payment: service is nil")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("payment: mux and handler are required")
	}
	user := func(next http.Handler) http.Handler { return authz.RequireUser(next) }
	mux.HandleFunc("GET /api/payment/methods", handler.methods)
	mux.Handle("POST /api/payment/methods", user(http.HandlerFunc(handler.createSubscription)))
	mux.Handle("POST /api/payment", user(http.HandlerFunc(handler.createPayment)))
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

func RegisterServiceRoutes(mux *http.ServeMux, service *Service) error {
	handler, err := NewHandler(service)
	if err != nil {
		return err
	}
	return RegisterRoutes(mux, handler)
}

type subscriptionPaymentRequest struct {
	PlanID string        `json:"planId"`
	Method PaymentMethod `json:"method"`
}

type paymentRequest struct {
	Amount   float64                    `json:"amount"`
	Currency string                     `json:"currency"`
	Method   PaymentMethod              `json:"method"`
	Metadata map[string]json.RawMessage `json:"metadata"`
}

type rechargeRequest struct {
	Amount  float64       `json:"amount"`
	Credits int           `json:"credits"`
	Method  PaymentMethod `json:"method"`
}

func (h *Handler) methods(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, h.service.AvailableMethods())
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	request, err := decodeJSON[subscriptionPaymentRequest](r)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	result, err := h.service.CreateSubscriptionPayment(r.Context(), principal.UserID, request.PlanID, request.Method)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, result)
}

func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	request, err := decodeJSON[paymentRequest](r)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	result, err := h.service.CreatePayment(r.Context(), CreatePaymentInput{UserID: principal.UserID, Amount: request.Amount, Currency: request.Currency, Method: request.Method, Metadata: request.Metadata})
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
	settings, err := h.service.Settings(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	if request.Amount < settings.MinRecharge {
		writePaymentError(w, r, appRuntime.NewAPIError(http.StatusBadRequest, "minimum_recharge", "recharge amount is below the minimum", nil))
		return
	}
	expectedCredits := int(request.Amount / settings.CreditPrice)
	if request.Credits != expectedCredits {
		writePaymentError(w, r, appRuntime.NewAPIError(http.StatusBadRequest, "invalid_credits", "credit calculation is invalid", nil))
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	metadata := map[string]json.RawMessage{"type": json.RawMessage(`"recharge"`), "credits": json.RawMessage(strconv.Itoa(request.Credits)), "creditPrice": json.RawMessage(strconv.FormatFloat(settings.CreditPrice, 'f', 2, 64))}
	result, err := h.service.CreatePayment(r.Context(), CreatePaymentInput{UserID: principal.UserID, Amount: request.Amount, Currency: "CNY", Method: request.Method, Metadata: metadata})
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, result)
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	principal, err := authz.RequestPrincipal(r.Context())
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
	h.respondOwnedPayment(w, r, func(outTradeNo string) (PaymentView, error) { return h.service.GetPayment(r.Context(), outTradeNo) })
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	h.respondOwnedPayment(w, r, func(outTradeNo string) (PaymentView, error) { return h.service.QueryPayment(r.Context(), outTradeNo) })
}

func (h *Handler) respondOwnedPayment(w http.ResponseWriter, r *http.Request, load func(string) (PaymentView, error)) {
	outTradeNo := r.PathValue("outTradeNo")
	view, err := load(outTradeNo)
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	if err := authz.CheckOwnership(principal, view.UserID); err != nil {
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
	principal, err := authz.RequestPrincipal(r.Context())
	if err != nil {
		writePaymentError(w, r, err)
		return
	}
	if err := authz.CheckOwnership(principal, view.UserID); err != nil {
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
