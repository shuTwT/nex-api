package mcpgateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/user"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	"github.com/shuTwT/nex-api/internal/service/stats"
)

var (
	ErrInsufficientCredits = errors.New("mcp gateway: insufficient credits")
	ErrInvalidService      = errors.New("mcp gateway: invalid MCP service configuration")
	ErrUnsupportedType     = errors.New("mcp gateway: unsupported MCP service type")
)

type CreditLedger interface {
	Reserve(ctx context.Context, userID, mcpID string, credits int) error
}

type AuditLogger interface {
	Record(ctx context.Context, entry serviceaccounts.AuditEntry) error
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

type NoopAudit struct{}

func (NoopAudit) Record(context.Context, serviceaccounts.AuditEntry) error { return nil }

type NoopStats struct{}

func (NoopStats) Increment(context.Context, stats.RequestEvent) error { return nil }
