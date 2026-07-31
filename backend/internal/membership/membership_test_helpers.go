package membership

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

func newMembershipClient(t *testing.T) (*ent.Client, time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "membership.db")
	database, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := database.ExecContext(context.Background(), membershipSchema); err != nil {
		t.Fatalf("create membership schema: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open ent: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client, time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
}

func seedMembershipUser(t *testing.T, client *ent.Client, id string, credits int) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	_, err := client.User.Create().SetID(id).SetName(id).SetEmail(id + "@example.com").SetUsername(id).SetPassword("redacted").SetRole("user").SetCredits(credits).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background())
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedPlan(t *testing.T, client *ent.Client, id, title string, credits, duration int, unit string) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	_, err := client.SubscriptionPlan.Create().SetID(id).SetTitle(title).SetPrice(10).SetTotalCredits(credits).SetSortOrder(0).SetValidityDuration(duration).SetValidityUnit(unit).SetCreditResetCycle("month").SetIsActive(true).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background())
	if err != nil {
		t.Fatalf("seed plan: %v", err)
	}
}

const membershipSchema = `
CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT, "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT, "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL DEFAULT 'user', "credits" INTEGER NOT NULL DEFAULT 1000, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "SubscriptionPlan" ("id" TEXT PRIMARY KEY NOT NULL, "title" TEXT NOT NULL UNIQUE, "price" REAL NOT NULL, "totalCredits" INTEGER NOT NULL, "sortOrder" INTEGER NOT NULL DEFAULT 0, "validityDuration" INTEGER NOT NULL, "validityUnit" TEXT NOT NULL DEFAULT 'day', "creditResetCycle" TEXT NOT NULL DEFAULT 'month', "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "Subscription" ("id" TEXT PRIMARY KEY NOT NULL, "planName" TEXT NOT NULL, "credits" INTEGER NOT NULL, "price" REAL NOT NULL, "startDate" DATETIME NOT NULL, "endDate" DATETIME NOT NULL, "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL, "paymentId" TEXT UNIQUE, "planId" TEXT, "userId" TEXT NOT NULL);
CREATE TABLE "RedemptionCode" ("id" TEXT PRIMARY KEY NOT NULL, "code" TEXT NOT NULL UNIQUE, "type" TEXT NOT NULL, "planId" TEXT, "planName" TEXT, "credits" INTEGER, "expiresAt" DATETIME, "isUsed" BOOLEAN NOT NULL DEFAULT 0, "usedBy" TEXT, "usedAt" DATETIME, "createdBy" TEXT NOT NULL, "batchId" TEXT, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "AuditLog" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT, "action" TEXT NOT NULL, "resource" TEXT NOT NULL, "details" TEXT, "ipAddress" TEXT, "userAgent" TEXT, "level" TEXT NOT NULL DEFAULT 'info', "status" TEXT NOT NULL DEFAULT 'success', "metadata" TEXT, "createdAt" DATETIME NOT NULL);
`
