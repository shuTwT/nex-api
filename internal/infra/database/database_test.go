package database

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/internal/infra/config"
)

func TestOpenSQLiteCreatesMissingParentDirectory(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "data", "dev.db")
	client, err := openSQLite(context.Background(), config.Database{
		ConnMaxLifetime: time.Hour,
	}, "file:"+databasePath, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	info, err := os.Stat(databasePath)
	if err != nil {
		t.Fatalf("stat created database: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("database path %q is a directory", databasePath)
	}
}
