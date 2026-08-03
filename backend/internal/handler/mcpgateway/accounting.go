package mcpgateway

import (
	"context"
	"log/slog"
	"net/http"

	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	"github.com/shuTwT/nex-api/backend/internal/service/authz"
	"github.com/shuTwT/nex-api/backend/internal/service/stats"
)

func (h *Handler) recordUsage(ctx context.Context, userID, identifier string, credits int, request *http.Request) {
	if err := h.services.Stats.Increment(ctx, stats.RequestEvent{UserID: userID, Alias: "mcp:" + identifier, Credits: float64(credits)}); err != nil {
		h.logger.WarnContext(ctx, "MCP Redis counter failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
	if err := h.services.Audits.Record(ctx, serviceaccounts.AuditEntry{UserID: userID, Action: "MCP call", Resource: "MCP: " + identifier, Details: "MCP service call", IPAddress: clientIP(request), UserAgent: request.UserAgent(), Level: "info", Status: "success"}); err != nil {
		h.logger.WarnContext(ctx, "MCP audit write failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
}

func (h *Handler) auditFailure(ctx context.Context, userID, identifier, details string, request *http.Request) {
	if err := h.services.Audits.Record(ctx, serviceaccounts.AuditEntry{UserID: userID, Action: "MCP call failed", Resource: "MCP: " + identifier, Details: details, IPAddress: clientIP(request), UserAgent: request.UserAgent(), Level: "warning", Status: "error"}); err != nil {
		h.logger.WarnContext(ctx, "MCP failure audit write failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
}

var _ Authenticator = (*authz.TokenService)(nil)
