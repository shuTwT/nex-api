package mcpgateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/user"
	"github.com/shuTwT/nex-api/backend/internal/stats"
)

type CreditLedger interface {
	Reserve(ctx context.Context, userID, mcpID string, credits int) error
}

type AuditLogger interface {
	Record(ctx context.Context, entry accounts.AuditEntry) error
}

type RequestCounter interface {
	Increment(ctx context.Context, event stats.RequestEvent) error
}

type EntCreditLedger struct {
	client *ent.Client
	now    func() time.Time
}

func NewEntCreditLedger(client *ent.Client) (*EntCreditLedger, error) {
	if client == nil {
		return nil, errors.New("mcp gateway: Ent client is nil")
	}
	return &EntCreditLedger{client: client, now: time.Now}, nil
}

func (s *EntCreditLedger) Reserve(ctx context.Context, userID, mcpID string, credits int) error {
	if ctx == nil || userID == "" || mcpID == "" || credits < 0 {
		return ErrInvalidService
	}
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("mcp gateway: begin credit reservation: %w", err)
	}
	updated, err := tx.User.Update().Where(user.ID(userID), user.CreditsGTE(credits)).AddCredits(-credits).Save(ctx)
	if err != nil {
		return abortCreditTx(tx, fmt.Errorf("mcp gateway: reserve credits: %w", err))
	}
	if updated != 1 {
		return abortCreditTx(tx, ErrInsufficientCredits)
	}
	if _, err := tx.McpUsage.Create().SetUserID(userID).SetMcpID(mcpID).SetCredits(credits).SetStatus("success").SetCreatedAt(s.now().UTC()).Save(ctx); err != nil {
		return abortCreditTx(tx, fmt.Errorf("mcp gateway: record MCP usage: %w", err))
	}
	if _, err := tx.McpService.UpdateOneID(mcpID).AddCallCount(1).Save(ctx); err != nil {
		return abortCreditTx(tx, fmt.Errorf("mcp gateway: increment MCP call count: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("mcp gateway: commit credit reservation: %w", err)
	}
	return nil
}

func abortCreditTx(tx *ent.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errors.Join(err, fmt.Errorf("mcp gateway: rollback credit reservation: %w", rollbackErr))
	}
	return err
}

type noopAudit struct{}

func (noopAudit) Record(context.Context, accounts.AuditEntry) error { return nil }

type noopStats struct{}

func (noopStats) Increment(context.Context, stats.RequestEvent) error { return nil }

func (h *Handler) recordUsage(ctx context.Context, userID, identifier string, credits int, request *http.Request) {
	if err := h.services.Stats.Increment(ctx, stats.RequestEvent{UserID: userID, Alias: "mcp:" + identifier, Credits: float64(credits)}); err != nil {
		h.logger.WarnContext(ctx, "MCP Redis counter failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
	if err := h.services.Audits.Record(ctx, accounts.AuditEntry{UserID: userID, Action: "MCP call", Resource: "MCP: " + identifier, Details: "MCP service call", IPAddress: clientIP(request), UserAgent: request.UserAgent(), Level: "info", Status: "success"}); err != nil {
		h.logger.WarnContext(ctx, "MCP audit write failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
}

func (h *Handler) auditFailure(ctx context.Context, userID, identifier, details string, request *http.Request) {
	if err := h.services.Audits.Record(ctx, accounts.AuditEntry{UserID: userID, Action: "MCP call failed", Resource: "MCP: " + identifier, Details: details, IPAddress: clientIP(request), UserAgent: request.UserAgent(), Level: "warning", Status: "error"}); err != nil {
		h.logger.WarnContext(ctx, "MCP failure audit write failed", slog.String("identifier", identifier), slog.Any("err", err))
	}
}

var _ Authenticator = (*authz.TokenService)(nil)
