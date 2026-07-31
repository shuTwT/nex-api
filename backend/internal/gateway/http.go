package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/httpapi/generated"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"github.com/shuTwT/nex-api/backend/internal/stats"
	"github.com/shuTwT/nex-api/backend/internal/worker"
)

type APIResolver interface {
	GetByIdentifier(context.Context, string) (*ent.Api, error)
}

type TokenAuthenticator interface {
	AuthenticateBearer(context.Context, string) (authz.Principal, error)
}

type Transformer interface {
	Execute(context.Context, worker.Job) (worker.TransformResult, error)
}

type UsageRecorder interface {
	Increment(context.Context, stats.RequestEvent) error
}

type Auditor interface {
	Record(context.Context, accounts.AuditEntry) error
}

type Options struct {
	APIs       APIResolver
	Tokens     TokenAuthenticator
	Transforms Transformer
	Usage      UsageRecorder
	Audit      Auditor
	Accountant Accountant
	Database   *ent.Client
	HTTPClient *http.Client
	Timeout    time.Duration
	Logger     *slog.Logger
	Now        func() time.Time
}

type Handler struct {
	apis       APIResolver
	tokens     TokenAuthenticator
	transforms Transformer
	usage      UsageRecorder
	audit      Auditor
	accounting Accountant
	client     *http.Client
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
		if options.Database == nil {
			return nil, errors.New("gateway: accountant or database is required")
		}
		accounting = NewEntAccountant(options.Database)
	}
	client := options.HTTPClient
	if client == nil {
		client = newHTTPClient()
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		apis: options.APIs, tokens: options.Tokens, transforms: options.Transforms,
		usage: options.Usage, audit: options.Audit, accounting: accounting,
		client: client, timeout: timeout, logger: logger,
		now: func() time.Time {
			if options.Now != nil {
				return options.Now()
			}
			return time.Now()
		},
	}, nil
}

func (h *Handler) V1AliasRouteGet(w http.ResponseWriter, r *http.Request, alias string, _ generated.V1AliasRouteGetParams) {
	h.serve(w, r, alias)
}

func (h *Handler) V1AliasRoutePost(w http.ResponseWriter, r *http.Request, alias string) {
	h.serve(w, r, alias)
}

func (h *Handler) V1AliasRoutePut(w http.ResponseWriter, r *http.Request, alias string) {
	h.serve(w, r, alias)
}

func (h *Handler) V1AliasRoutePatch(w http.ResponseWriter, r *http.Request, alias string) {
	h.serve(w, r, alias)
}

func (h *Handler) V1AliasRouteDelete(w http.ResponseWriter, r *http.Request, alias string) {
	h.serve(w, r, alias)
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, alias string) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	principal, err := h.tokens.AuthenticateBearer(ctx, r.Header.Get("Authorization"))
	if err != nil {
		h.fail(w, r, alias, "", http.StatusUnauthorized, "unauthorized", "invalid API token")
		return
	}
	api, err := h.apis.GetByIdentifier(ctx, alias)
	if err != nil {
		status := http.StatusInternalServerError
		code, message := "internal_error", "internal server error"
		if apiErr := runtime.FromError(err); apiErr.StatusCode == http.StatusNotFound {
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
	reservation, err := h.accounting.Reserve(ctx, CreditRequest{UserID: principal.UserID, APIID: api.ID, Credits: api.Pricing})
	if err != nil {
		status, code, message := http.StatusInternalServerError, "internal_error", "credit reservation failed"
		if errors.Is(err, ErrInsufficientCredits) {
			status, code, message = http.StatusPaymentRequired, "insufficient_credits", "insufficient credits"
		}
		h.fail(w, r, api.Name, principal.UserID, status, code, message)
		return
	}

	requestData := transformRequest(r, body)
	if api.PreScript != "" {
		requestData, err = h.preTransform(ctx, api.PreScript, requestData)
		if err != nil {
			h.refund(ctx, reservation, r, api.Name, principal.UserID, "pre-transform failed")
			h.writeError(w, http.StatusInternalServerError, "transform_error", "request transform failed")
			return
		}
	}
	response, err := h.forward(ctx, r, api.Endpoint, requestData)
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
		response, err = h.postTransform(ctx, api.PostScript, response)
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

func (h *Handler) preTransform(ctx context.Context, script string, input requestData) (requestData, error) {
	if h.transforms == nil {
		return requestData{}, errors.New("gateway: transform worker is unavailable")
	}
	result, err := h.transforms.Execute(ctx, worker.Job{ID: uuid.NewString(), Kind: worker.ScriptKindPre, Script: script, Headers: input.Headers, Query: input.Query, Body: transformPayload(input.Body, input.ContentType)})
	if err != nil {
		return requestData{}, fmt.Errorf("pre-transform: %w", err)
	}
	mergeStrings(input.Headers, result.Headers)
	mergeQuery(input.Query, result.Query)
	input.QueryChanged = true
	if result.BodySet {
		input.Body = transformedBytes(result.Body, input.ContentType)
	}
	return input, nil
}

func (h *Handler) postTransform(ctx context.Context, script string, response proxyResponse) (proxyResponse, error) {
	if h.transforms == nil {
		return proxyResponse{}, errors.New("gateway: transform worker is unavailable")
	}
	result, err := h.transforms.Execute(ctx, worker.Job{ID: uuid.NewString(), Kind: worker.ScriptKindPost, Script: script, ResponseBody: response.TransformBody, ResponseHeaders: response.TransformHeaders})
	if err != nil {
		return proxyResponse{}, fmt.Errorf("post-transform: %w", err)
	}
	for key, value := range result.ResponseHeaders {
		response.Headers.Del(key)
		response.Headers.Set(key, value)
	}
	if result.ResponseBodySet {
		response.Body = transformedBytes(result.ResponseBody, response.ContentType)
	}
	return response, nil
}

func (h *Handler) refund(ctx context.Context, reservation CreditReservation, r *http.Request, resource, userID, details string) {
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
	err := h.audit.Record(ctx, accounts.AuditEntry{UserID: userID, Action: "API call", Resource: resource, Details: details, IPAddress: clientIP(r), UserAgent: r.UserAgent(), Level: "info", Status: status})
	if err != nil {
		h.logger.ErrorContext(ctx, "gateway audit write failed", slog.Any("err", err), slog.String("resource", resource))
	}
}
