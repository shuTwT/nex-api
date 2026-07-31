package mcpgateway

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/accounts"
	"github.com/shuTwT/nex-api/backend/internal/authz"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/stats"
)

type fakeAuth struct{}

func (fakeAuth) AuthenticateBearer(context.Context, string) (authz.Principal, error) {
	return authz.Principal{UserID: "user-1", TokenID: "token-1", Source: authz.APITokenCredential}, nil
}

type fakeCatalog struct{ service *ent.McpService }

func (c fakeCatalog) GetByIdentifier(context.Context, string) (*ent.McpService, error) {
	return c.service, nil
}

type fakeCredits struct {
	mu          sync.Mutex
	calls       int
	lastCredits int
	err         error
}

func (c *fakeCredits) Reserve(_ context.Context, _ string, _ string, credits int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	c.lastCredits = credits
	return c.err
}

type fakeStats struct{ events []stats.RequestEvent }

func (s *fakeStats) Increment(_ context.Context, event stats.RequestEvent) error {
	s.events = append(s.events, event)
	return nil
}

type fakeAudit struct{ entries []accounts.AuditEntry }

func (a *fakeAudit) Record(_ context.Context, entry accounts.AuditEntry) error {
	a.entries = append(a.entries, entry)
	return nil
}

func newTestHandler(t *testing.T, service *ent.McpService) *Handler {
	t.Helper()
	return newTestHandlerWithServices(t, service, &fakeCredits{}, &fakeStats{})
}

func newTestHandlerWithServices(t *testing.T, service *ent.McpService, credits CreditLedger, requestStats RequestCounter, audits ...AuditLogger) *Handler {
	t.Helper()
	var audit AuditLogger
	if len(audits) > 0 {
		audit = audits[0]
	}
	handler, err := New(Services{
		Authenticator: fakeAuth{},
		Catalog:       fakeCatalog{service: service},
		Credits:       credits,
		Audits:        audit,
		Stats:         requestStats,
	}, HandlerOptions{
		MaxRequestBytes:  1 << 20,
		MaxResponseBytes: 1 << 20,
		MaxRuntime:       2 * time.Second,
		Stdio: StdioOptions{
			MaxRuntime:     2 * time.Second,
			MaxMemoryBytes: 64 << 20,
			MaxOutputBytes: 1 << 20,
			MaxFrameBytes:  1 << 20,
		},
	})
	requireNoError(t, err)
	return handler
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
