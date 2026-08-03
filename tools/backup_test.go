package tools

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBackupRestore(t *testing.T) {
	// Given
	ctx := context.Background()
	root := t.TempDir()
	sourceDB := filepath.Join(root, "prisma", "dev.db")
	sourceUploads := filepath.Join(root, "data", "upload")
	if err := os.MkdirAll(filepath.Dir(sourceDB), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sourceUploads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := createLegacyDatabase(sourceDB); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceUploads, "avatar.txt"), []byte("avatar"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(sourceUploads, "avatar.txt"), filepath.Join(sourceUploads, "avatar-link.txt")); err != nil {
		t.Fatal(err)
	}
	beforeHash, err := fileSHA256(sourceDB)
	if err != nil {
		t.Fatal(err)
	}

	// When
	backupDir := filepath.Join(root, "backup")
	manifest, err := CreateBackup(ctx, BackupConfig{
		SourceDatabase: sourceDB,
		SourceUploads:  sourceUploads,
		Destination:    backupDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMigrations(ctx, MigrationConfig{
		DatabasePath:  manifest.DatabasePath,
		MigrationsDir: migrationsDir(t),
		Baseline:      "20260730160000",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyDatabase(ctx, manifest.DatabasePath); err != nil {
		t.Fatal(err)
	}
	rollbackDir := filepath.Join(root, "rollback")
	if err := RestoreBackup(ctx, RestoreConfig{
		BackupDirectory: backupDir,
		Destination:     rollbackDir,
	}); err != nil {
		t.Fatal(err)
	}

	// Then
	if manifest.RowCounts["User"] != 1 || len(manifest.RowCounts) != len(coreTables) {
		t.Fatalf("unexpected manifest row counts: %#v", manifest.RowCounts)
	}
	if !manifest.ForeignKeysOK || len(manifest.Uploads) != 1 {
		t.Fatalf("unexpected manifest integrity: %#v", manifest)
	}
	if _, err := os.Stat(filepath.Join(rollbackDir, "upload", "avatar-link.txt")); !os.IsNotExist(err) {
		t.Fatalf("symlink was restored: %v", err)
	}
	if got, err := fileSHA256(sourceDB); err != nil || got != beforeHash {
		t.Fatalf("source database changed: got %s want %s err %v", got, beforeHash, err)
	}
	if got, err := os.ReadFile(filepath.Join(rollbackDir, "upload", "avatar.txt")); err != nil || string(got) != "avatar" {
		t.Fatalf("restored upload mismatch: %q err %v", got, err)
	}
}

func createLegacyDatabase(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()
	migrationPath := filepath.Join(repoBackendDir(), "migrations", "20260730160000_legacy_baseline.sql")
	schema, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO "User" (id, email, username, password, role, credits, updatedAt)
		VALUES ('user-1', 'user@example.com', 'user1', 'hash', 'user', 1000, CURRENT_TIMESTAMP);
		INSERT INTO "SubscriptionPlan" (id, title, price, totalCredits, validityDuration, updatedAt)
		VALUES ('plan-1', 'Starter', 1.5, 100, 30, CURRENT_TIMESTAMP);
		INSERT INTO "Payment" (id, userId, outTradeNo, method, amount, updatedAt)
		VALUES ('payment-1', 'user-1', 'order-1', 'mock', 1.5, CURRENT_TIMESTAMP);
		INSERT INTO "Subscription" (id, userId, planId, planName, credits, price, endDate, paymentId, updatedAt)
		VALUES ('subscription-1', 'user-1', 'plan-1', 'Starter', 100, 1.5, CURRENT_TIMESTAMP, 'payment-1', CURRENT_TIMESTAMP);
		INSERT INTO "ApiToken" (id, userId, name, token, permissions, updatedAt)
		VALUES ('token-1', 'user-1', 'test', 'token-value', '{}', CURRENT_TIMESTAMP);
	`)
	return err
}

func migrationsDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "migrations")
}

func repoBackendDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
