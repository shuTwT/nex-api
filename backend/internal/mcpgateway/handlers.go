package mcpgateway

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

const (
	defaultRequestBytes  = 1 << 20
	defaultResponseBytes = 8 << 20
)

var (
	ErrInsufficientCredits = errors.New("mcp gateway: insufficient credits")
	ErrInvalidService      = errors.New("mcp gateway: invalid MCP service configuration")
	ErrUnsupportedType     = errors.New("mcp gateway: unsupported MCP service type")
)

type Authenticator interface {
	AuthenticateBearer(ctx context.Context, header string) (authz.Principal, error)
}

type Catalog interface {
	GetByIdentifier(ctx context.Context, identifier string) (*ent.McpService, error)
}

type Services struct {
	Authenticator Authenticator
	Catalog       Catalog
	Credits       CreditLedger
	Audits        AuditLogger
	Stats         RequestCounter
}

type HandlerOptions struct {
	MaxRequestBytes  int64
	MaxResponseBytes int64
	MaxRuntime       time.Duration
	HTTPClient       *http.Client
	Stdio            StdioOptions
	Logger           *slog.Logger
}

type Handler struct {
	services Services
	options  HandlerOptions
	stdio    *StdioRunner
	logger   *slog.Logger
	mux      *http.ServeMux
}

func New(services Services, options HandlerOptions) (*Handler, error) {
	if services.Authenticator == nil || services.Catalog == nil || services.Credits == nil {
		return nil, errors.New("mcp gateway: required services are incomplete")
	}
	if services.Audits == nil {
		services.Audits = noopAudit{}
	}
	if services.Stats == nil {
		services.Stats = noopStats{}
	}
	options = normalizeHandlerOptions(options)
	stdio, err := NewStdioRunner(options.Stdio)
	if err != nil {
		return nil, err
	}
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{}
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	handler := &Handler{services: services, options: options, stdio: stdio, logger: options.Logger, mux: http.NewServeMux()}
	handler.mux.HandleFunc("OPTIONS /api/v1/mcp/{identifier}", handler.optionsRoute)
	handler.mux.HandleFunc("POST /api/v1/mcp/{identifier}", handler.postRoute)
	return handler, nil
}

func NewHandler(services Services, options ...HandlerOptions) http.Handler {
	var handlerOptions HandlerOptions
	if len(options) > 0 {
		handlerOptions = options[0]
	}
	handler, err := New(services, handlerOptions)
	if err != nil {
		return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writeError(writer, http.StatusInternalServerError, "mcp_gateway_unavailable")
		})
	}
	return handler
}

func RegisterRoutes(mux *http.ServeMux, services Services, options ...HandlerOptions) error {
	if mux == nil {
		return errors.New("mcp gateway: route mux is nil")
	}
	var handlerOptions HandlerOptions
	if len(options) > 0 {
		handlerOptions = options[0]
	}
	handler, err := New(services, handlerOptions)
	if err != nil {
		return err
	}
	mux.Handle("/api/v1/mcp/", handler)
	return nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.mux.ServeHTTP(writer, request)
}

func (h *Handler) V1McpIdentifierRouteOptions(writer http.ResponseWriter, request *http.Request, _ string) {
	h.optionsRoute(writer, request)
}

func (h *Handler) V1McpIdentifierRoutePost(writer http.ResponseWriter, request *http.Request, identifier string) {
	h.postIdentifier(writer, request, identifier)
}

func (h *Handler) optionsRoute(writer http.ResponseWriter, _ *http.Request) {
	setCORSHeaders(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (h *Handler) postRoute(writer http.ResponseWriter, request *http.Request) {
	h.postIdentifier(writer, request, request.PathValue("identifier"))
}

func (h *Handler) postIdentifier(writer http.ResponseWriter, request *http.Request, identifier string) {
	setCORSHeaders(writer)
	ctx := request.Context()
	principal, err := h.services.Authenticator.AuthenticateBearer(ctx, request.Header.Get("Authorization"))
	if err != nil {
		h.auditFailure(ctx, "", identifier, "invalid API token", request)
		writeError(writer, http.StatusUnauthorized, "invalid_api_token")
		return
	}
	service, err := h.services.Catalog.GetByIdentifier(ctx, identifier)
	if err != nil {
		if ent.IsNotFound(err) {
			h.auditFailure(ctx, principal.UserID, identifier, "MCP service not found", request)
			writeError(writer, http.StatusNotFound, "mcp_service_not_found")
			return
		}
		h.auditFailure(ctx, principal.UserID, identifier, "MCP service lookup failed", request)
		writeError(writer, http.StatusInternalServerError, "mcp_service_lookup_failed")
		return
	}
	if service == nil {
		writeError(writer, http.StatusNotFound, "mcp_service_not_found")
		return
	}
	if !service.IsActive {
		writeError(writer, http.StatusForbidden, "mcp_service_disabled")
		return
	}
	if service.Pricing < 0 {
		writeError(writer, http.StatusInternalServerError, "invalid_mcp_service")
		return
	}
	envVars, err := h.validateService(service)
	if err != nil {
		if errors.Is(err, ErrUnsupportedType) {
			writeError(writer, http.StatusBadRequest, "unsupported_mcp_service_type")
			return
		}
		writeError(writer, http.StatusInternalServerError, "invalid_mcp_service")
		return
	}
	body, err := readRequestBody(request, h.options.MaxRequestBytes)
	if err != nil {
		writeError(writer, http.StatusRequestEntityTooLarge, "request_body_too_large")
		return
	}
	if err := h.services.Credits.Reserve(ctx, principal.UserID, service.ID, service.Pricing); err != nil {
		if errors.Is(err, ErrInsufficientCredits) {
			writeError(writer, http.StatusPaymentRequired, "insufficient_credits")
			return
		}
		h.auditFailure(ctx, principal.UserID, service.Identifier, "credit reservation failed", request)
		writeError(writer, http.StatusInternalServerError, "credit_reservation_failed")
		return
	}
	h.recordUsage(ctx, principal.UserID, service.Identifier, service.Pricing, request)
	callContext, cancel := context.WithTimeout(ctx, h.options.MaxRuntime)
	defer cancel()
	switch service.Type {
	case "stdio":
		h.serveStdio(writer, callContext, service, envVars, body)
	case "sse":
		h.serveSSE(writer, callContext, request, service, body)
	case "streamableHttp":
		h.serveStreamable(writer, callContext, request, service, body)
	default:
		writeError(writer, http.StatusBadRequest, "unsupported_mcp_service_type")
	}
}

func normalizeHandlerOptions(options HandlerOptions) HandlerOptions {
	if options.MaxRequestBytes <= 0 {
		options.MaxRequestBytes = defaultRequestBytes
	}
	if options.MaxResponseBytes <= 0 {
		options.MaxResponseBytes = defaultResponseBytes
	}
	if options.MaxRuntime <= 0 {
		options.MaxRuntime = 30 * time.Second
	}
	options.Stdio = normalizeStdioOptions(options.Stdio)
	return options
}

var _ http.Handler = (*Handler)(nil)
