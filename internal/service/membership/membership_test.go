package membership

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/subscription"
)

func TestCalculateEndDate_usesCalendarUnits(t *testing.T) {
	for _, test := range []struct {
		name, unit  string
		start, want time.Time
		duration    int
	}{
		{"day", "day", time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), time.Date(2024, 2, 2, 12, 0, 0, 0, time.UTC), 2},
		{"week", "week", time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), time.Date(2024, 2, 14, 12, 0, 0, 0, time.UTC), 2},
		{"month", "month", time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC), 1},
		{"year", "year", time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC), time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC), 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculateEndDate(test.start, test.unit, test.duration)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(test.want) {
				t.Fatalf("end date = %s, want %s", got, test.want)
			}
		})
	}
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestMembershipService_subscribe_concurrentlyLeavesOneActiveSubscription(t *testing.T) {
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	seedPlan(t, client, "plan-1", "Starter", 25, 1, "month")
	service, err := NewMembershipService(client, fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	const attempts = 12
	errs := make(chan error, attempts)
	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Subscribe(context.Background(), "user-1", "plan-1")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent subscribe: %v", err)
		}
	}
	active, err := client.Subscription.Query().Where(subscription.UserId("user-1"), subscription.IsActive(true)).Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if active != 1 {
		t.Fatalf("active subscriptions = %d", active)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 100+attempts*25 {
		t.Fatalf("credits = %d", user.Credits)
	}
	audits, err := client.AuditLog.Query().Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if audits != attempts {
		t.Fatalf("audit events = %d", audits)
	}
}

func TestRedemptionService_redeem_concurrentlyClaimsCodeOnce(t *testing.T) {
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	if _, err := client.RedemptionCode.Create().SetID("code-1").SetCode("ONCEONLY").SetType("quota").SetCredits(50).SetCreatedBy("admin").SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	service, err := NewRedemptionService(client, fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	const attempts = 12
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Redeem(context.Background(), "user-1", "onceonly")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful redemptions = %d", successes)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 150 {
		t.Fatalf("credits = %d", user.Credits)
	}
	code, err := client.RedemptionCode.Get(context.Background(), "code-1")
	if err != nil {
		t.Fatal(err)
	}
	if !code.IsUsed || code.UsedBy != "user-1" {
		t.Fatalf("code usage = %+v", code)
	}
	audits, err := client.AuditLog.Query().Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if audits != 1 {
		t.Fatalf("audit events = %d", audits)
	}
}

func TestRedemptionService_subscriptionRedeemGrantsPlanCredits(t *testing.T) {
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	seedPlan(t, client, "plan-1", "Starter", 25, 1, "month")
	if _, err := client.RedemptionCode.Create().SetID("code-1").SetCode("SUBONLY").SetType("subscription").SetPlanId("plan-1").SetPlanName("Starter").SetCreatedBy("admin").SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	service, err := NewRedemptionService(client, fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Redeem(context.Background(), "user-1", "subonly")
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "subscription" {
		t.Fatalf("result type = %q", result.Type)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Credits != 125 {
		t.Fatalf("credits = %d", user.Credits)
	}
	active, err := client.Subscription.Query().Where(subscription.UserId("user-1"), subscription.IsActive(true)).Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if active != 1 {
		t.Fatalf("active subscriptions = %d", active)
	}
}

func newMembershipClient(t *testing.T) (*ent.Client, time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "membership.db")
	database, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(context.Background(), membershipSchema); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	client, err := ent.Open(dialect.SQLite, path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client, time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
}
func seedMembershipUser(t *testing.T, client *ent.Client, id string, credits int) {
	t.Helper()
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	if _, err := client.User.Create().SetID(id).SetName(id).SetEmail(id + "@example.com").SetUsername(id).SetPassword("redacted").SetRole("user").SetCredits(credits).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
}
func seedPlan(t *testing.T, client *ent.Client, id, title string, credits, duration int, unit string) {
	t.Helper()
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	if _, err := client.SubscriptionPlan.Create().SetID(id).SetTitle(title).SetPrice(10).SetTotalCredits(credits).SetSortOrder(0).SetValidityDuration(duration).SetValidityUnit(unit).SetCreditResetCycle("month").SetIsActive(true).SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
}

const membershipSchema = `
CREATE TABLE "User" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT, "email" TEXT NOT NULL, "emailVerified" DATETIME, "image" TEXT, "username" TEXT NOT NULL, "password" TEXT NOT NULL, "role" TEXT NOT NULL DEFAULT 'user', "credits" INTEGER NOT NULL DEFAULT 1000, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "SubscriptionPlan" ("id" TEXT PRIMARY KEY NOT NULL, "title" TEXT NOT NULL UNIQUE, "price" REAL NOT NULL, "totalCredits" INTEGER NOT NULL, "sortOrder" INTEGER NOT NULL DEFAULT 0, "validityDuration" INTEGER NOT NULL, "validityUnit" TEXT NOT NULL DEFAULT 'day', "creditResetCycle" TEXT NOT NULL DEFAULT 'month', "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "Subscription" ("id" TEXT PRIMARY KEY NOT NULL, "planName" TEXT NOT NULL, "credits" INTEGER NOT NULL, "price" REAL NOT NULL, "startDate" DATETIME NOT NULL, "endDate" DATETIME NOT NULL, "isActive" BOOLEAN NOT NULL DEFAULT 1, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL, "paymentId" TEXT UNIQUE, "planId" TEXT, "userId" TEXT NOT NULL);
CREATE TABLE "RedemptionCode" ("id" TEXT PRIMARY KEY NOT NULL, "code" TEXT NOT NULL UNIQUE, "type" TEXT NOT NULL, "planId" TEXT, "planName" TEXT, "credits" INTEGER, "expiresAt" DATETIME, "isUsed" BOOLEAN NOT NULL DEFAULT 0, "usedBy" TEXT, "usedAt" DATETIME, "createdBy" TEXT NOT NULL, "batchId" TEXT, "createdAt" DATETIME NOT NULL, "updatedAt" DATETIME NOT NULL);
CREATE TABLE "AuditLog" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT, "action" TEXT NOT NULL, "resource" TEXT NOT NULL, "details" TEXT, "ipAddress" TEXT, "userAgent" TEXT, "level" TEXT NOT NULL DEFAULT 'info', "status" TEXT NOT NULL DEFAULT 'success', "metadata" TEXT, "createdAt" DATETIME NOT NULL);`
