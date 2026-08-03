package dashboard

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
)

func TestMonthlyCreditsUsedReturnsZeroWhenNoUsageExists(t *testing.T) {
	ctx := context.Background()
	client, err := ent.Open(dialect.SQLite, "file:"+filepath.Join(t.TempDir(), "dashboard.db")+"?_fk=1")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	credits, err := monthlyCreditsUsed(ctx, client, "new-user", time.Now())
	if err != nil {
		t.Fatalf("sum monthly credits: %v", err)
	}
	if credits != 0 {
		t.Fatalf("monthly credits = %d, want 0", credits)
	}
}
