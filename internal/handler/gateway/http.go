package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	servicegateway "github.com/shuTwT/nex-api/internal/service/gateway"
	"github.com/shuTwT/nex-api/internal/service/stats"
)

type APIResolver interface {
	GatewayAPI(context.Context, string) (servicecatalog.GatewayAPI, error)
}

type TokenAuthenticator interface {
	AuthenticateBearer(context.Context, string) (serviceauthz.Principal, error)
}

type UsageRecorder interface {
	Increment(context.Context, stats.RequestEvent) error
}

type Auditor interface {
	Record(context.Context, serviceaccounts.AuditEntry) error
}

type Options struct {
	APIs       APIResolver
	Tokens     TokenAuthenticator
	Transforms servicegateway.Transformer
	Usage      UsageRecorder
	Audit      Auditor
	Accountant servicegateway.Accountant
	Forwarder  servicegateway.Forwarder
	Timeout    time.Duration
	Logger     *slog.Logger
	Now        func() time.Time
}

type Handler struct {
	apis       APIResolver
	tokens     TokenAuthenticator
	transforms *servicegateway.TransformService
	usage      UsageRecorder
	audit      Auditor
	accounting servicegateway.Accountant
	forwarder  servicegateway.Forwarder
	timeout    time.Duration
	logger     *slog.Logger
	now        func() time.Time
}

func New(options Options) (*Handler, error) {
	if options.APIs == nil || options.Tokens == nil {
		return nil, errors.New("gateway: APIs and tokens are required")
	}
	accounting := options.Accountant
	if accounting == nil {
		return nil, errors.New("gateway: accountant is required")
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	transforms := servicegateway.NewTransformService(options.Transforms)
	forwarder := options.Forwarder
	if forwarder == nil {
		forwarder = servicegateway.NewProxyForwarder()
	}
	return &Handler{
		apis: options.APIs, tokens: options.Tokens, transforms: transforms,
		usage: options.Usage, audit: options.Audit, accounting: accounting, forwarder: forwarder,
		timeout: timeout, logger: logger,
		now: func() time.Time {
			if options.Now != nil {
				return options.Now()
			}
			return time.Now()
		},
	}, nil
}

// RegisterRoutes exposes the API gateway's method-agnostic proxy route.
func (h *Handler) RegisterRoutes(router chi.Router) error {
	if router == nil {
		return errors.New("gateway: route router is nil")
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		router.Method(method, "/api/v1/{alias}", http.HandlerFunc(h.aliasRoute))
	}
	return nil
}

func (h *Handler) aliasRoute(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, chi.URLParam(r, "alias"))
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, alias string) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	principal, err := h.tokens.AuthenticateBearer(ctx, r.Header.Get("Authorization"))
	if err != nil {
		h.fail(w, r, alias, "", http.StatusUnauthorized, "unauthorized", "invalid API token")
		return
	}
	api, err := h.apis.GatewayAPI(ctx, alias)
	if err != nil {
		status := http.StatusInternalServerError
		code, message := "internal_error", "internal server error"
		if apiErr := appRuntime.FromError(err); apiErr.StatusCode == http.StatusNotFound {
			status, code, message = http.StatusNotFound, "not_found", "API not found"
		}
		h.fail(w, r, alias, principal.UserID, status, code, message)
		return
	}
	if !api.IsActive {
		h.fail(w, r, api.Name, principal.UserID, http.StatusForbidden, "forbidden", "API is disabled")
		return
	}
	if !methodAllowed(api.Method, r.Method) {
		w.Header().Set("Allow", api.Method)
		h.fail(w, r, api.Name, principal.UserID, http.StatusMethodNotAllowed, "method_not_allowed", "request method is not supported")
		return
	}

	body, err := readBody(r)
	if err != nil {
		h.fail(w, r, api.Name, principal.UserID, http.StatusBadRequest, "invalid_body", "request body could not be read")
		return
	}
	reservation, err := h.accounting.Reserve(ctx, servicegateway.CreditRequest{UserID: principal.UserID, APIID: api.ID, Credits: api.Pricing})
	if err != nil {
		status, code, message := http.StatusInternalServerError, "internal_error", "credit reservation failed"
		if errors.Is(err, servicegateway.ErrInsufficientCredits) {
			status, code, message = http.StatusPaymentRequired, "insufficient_credits", "insufficient credits"
		}
		h.fail(w, r, api.Name, principal.UserID, status, code, message)
		return
	}

	requestData := servicegateway.TransformRequest(r.Header, r.URL.Query(), body)
	if api.PreScript != "" {
		requestData, err = h.transforms.PreTransform(ctx, api.PreScript, requestData)
		if err != nil {
			h.refund(ctx, reservation, r, api.Name, principal.UserID, "pre-transform failed")
			h.writeError(w, http.StatusInternalServerError, "transform_error", "request transform failed")
			return
		}
	}
	response, err := h.forwarder.Forward(ctx, api.Endpoint, servicegateway.ForwardRequest{Method: r.Method, Headers: r.Header, Query: r.URL.Query(), Body: body, Data: requestData})
	if err != nil {
		h.refund(ctx, reservation, r, api.Name, principal.UserID, "upstream request failed")
		status, code, message := http.StatusBadGateway, "upstream_error", "upstream request failed"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			status, code, message = http.StatusGatewayTimeout, "upstream_timeout", "upstream request timed out"
		}
		h.writeError(w, status, code, message)
		return
	}
	if api.PostScript != "" {
		response, err = h.transforms.PostTransform(ctx, api.PostScript, response)
		if err != nil {
			h.refund(ctx, reservation, r, api.Name, principal.UserID, "post-transform failed")
			h.writeError(w, http.StatusInternalServerError, "transform_error", "response transform failed")
			return
		}
	}
	if err := h.accounting.Finalize(ctx, reservation); err != nil {
		h.refund(ctx, reservation, r, api.Name, principal.UserID, "accounting finalization failed")
		h.logger.ErrorContext(ctx, "gateway accounting finalize failed", slog.Any("err", err), slog.String("reservation_id", reservation.ID))
		h.writeError(w, http.StatusInternalServerError, "accounting_error", "request accounting failed")
		return
	}
	if h.usage != nil {
		if err := h.usage.Increment(ctx, stats.RequestEvent{UserID: principal.UserID, Alias: alias, Credits: float64(api.Pricing), At: h.now()}); err != nil {
			h.logger.ErrorContext(ctx, "gateway usage accounting failed", slog.Any("err", err), slog.String("alias", alias))
		}
	}
	h.recordAudit(ctx, r, principal.UserID, api.Name, fmt.Sprintf("completed upstream response with status %d", response.StatusCode), "success")
	writeResponse(w, response)
}

func (h *Handler) refund(ctx context.Context, reservation servicegateway.CreditReservation, r *http.Request, resource, userID, details string) {
	refundCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), h.timeout)
	defer cancel()
	if err := h.accounting.Refund(refundCtx, reservation); err != nil {
		h.logger.ErrorContext(refundCtx, "gateway credit refund failed", slog.Any("err", err), slog.String("reservation_id", reservation.ID))
	}
	h.recordAudit(refundCtx, r, userID, resource, details, "error")
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, resource, userID string, status int, code, message string) {
	h.recordAudit(r.Context(), r, userID, resource, message, "error")
	h.writeError(w, status, code, message)
}

func (h *Handler) recordAudit(ctx context.Context, r *http.Request, userID, resource, details, status string) {
	if h.audit == nil {
		return
	}
	err := h.audit.Record(ctx, serviceaccounts.AuditEntry{UserID: userID, Action: "API call", Resource: resource, Details: details, IPAddress: clientIP(r), UserAgent: r.UserAgent(), Level: "info", Status: status})
	if err != nil {
		h.logger.ErrorContext(ctx, "gateway audit write failed", slog.Any("err", err), slog.String("resource", resource))
	}
}
