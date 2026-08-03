package mcpgateway

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	serviceerror "github.com/shuTwT/nex-api/backend/internal/service/apierror"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
	servicemcpgateway "github.com/shuTwT/nex-api/backend/internal/service/mcpgateway"
)

const (
	defaultRequestBytes  = 1 << 20
	defaultResponseBytes = 8 << 20
)

type Authenticator interface {
	AuthenticateBearer(ctx context.Context, header string) (serviceauthz.Principal, error)
}

type Catalog interface {
	GatewayMCPService(ctx context.Context, identifier string) (servicecatalog.GatewayMCPService, error)
}

type Services struct {
	Authenticator Authenticator
	Catalog       Catalog
	Credits       servicemcpgateway.CreditLedger
	Audits        servicemcpgateway.AuditLogger
	Stats         servicemcpgateway.RequestCounter
}

type HandlerOptions struct {
	MaxRequestBytes  int64
	MaxResponseBytes int64
	MaxRuntime       time.Duration
	Stdio            servicemcpgateway.StdioOptions
	Executor         servicemcpgateway.Executor
	Logger           *slog.Logger
}

type Handler struct {
	services Services
	options  HandlerOptions
	executor servicemcpgateway.Executor
	logger   *slog.Logger
	router   chi.Router
}

func New(services Services, options HandlerOptions) (*Handler, error) {
	if services.Authenticator == nil || services.Catalog == nil || services.Credits == nil {
		return nil, errors.New("mcp gateway: required services are incomplete")
	}
	if services.Audits == nil {
		services.Audits = servicemcpgateway.NoopAudit{}
	}
	if services.Stats == nil {
		services.Stats = servicemcpgateway.NoopStats{}
	}
	options = normalizeHandlerOptions(options)
	executor := options.Executor
	if executor == nil {
		var err error
		executor, err = servicemcpgateway.NewProxyExecutor(options.Stdio)
		if err != nil {
			return nil, err
		}
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	handler := &Handler{services: services, options: options, executor: executor, logger: options.Logger, router: chi.NewRouter()}
	handler.router.Options("/api/v1/mcp/{identifier}", handler.optionsRoute)
	handler.router.Post("/api/v1/mcp/{identifier}", handler.postRoute)
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

func RegisterRoutes(router chi.Router, services Services, options ...HandlerOptions) error {
	if router == nil {
		return errors.New("mcp gateway: route router is nil")
	}
	var handlerOptions HandlerOptions
	if len(options) > 0 {
		handlerOptions = options[0]
	}
	handler, err := New(services, handlerOptions)
	if err != nil {
		return err
	}
	router.Options("/api/v1/mcp/{identifier}", handler.optionsRoute)
	router.Post("/api/v1/mcp/{identifier}", handler.postRoute)
	return nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.router.ServeHTTP(writer, request)
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
	identifier := request.PathValue("identifier")
	if identifier == "" {
		identifier = chi.URLParam(request, "identifier")
		request.SetPathValue("identifier", identifier)
	}
	h.postIdentifier(writer, request, identifier)
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
	service, err := h.services.Catalog.GatewayMCPService(ctx, identifier)
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			h.auditFailure(ctx, principal.UserID, identifier, "MCP service not found", request)
			writeError(writer, http.StatusNotFound, "mcp_service_not_found")
			return
		}
		h.auditFailure(ctx, principal.UserID, identifier, "MCP service lookup failed", request)
		writeError(writer, http.StatusInternalServerError, "mcp_service_lookup_failed")
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
	if err := h.executor.Validate(service); err != nil {
		if errors.Is(err, servicemcpgateway.ErrUnsupportedType) {
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
		if errors.Is(err, servicemcpgateway.ErrInsufficientCredits) {
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
	response, err := h.executor.Invoke(callContext, service, servicemcpgateway.Invocation{Headers: request.Header, Body: body})
	if err != nil {
		writeError(writer, http.StatusBadGateway, "mcp_upstream_failed")
		return
	}
	defer response.Body.Close()
	for key, values := range response.Headers {
		writer.Header().Del(key)
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}
	writer.Header().Del("Content-Length")
	writer.WriteHeader(response.StatusCode)
	if err := writeStream(writer, response.Body, h.options.MaxResponseBytes); err != nil {
		h.logger.WarnContext(ctx, "MCP stream failed", slog.String("identifier", service.Identifier), slog.Any("err", err))
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
	return options
}

var _ http.Handler = (*Handler)(nil)
