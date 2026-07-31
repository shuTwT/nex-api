package membership

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/subscription"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestCalculateEndDate_usesCalendarUnits(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		unit     string
		duration int
		want     time.Time
	}{
		{name: "day", start: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), unit: "day", duration: 2, want: time.Date(2024, 2, 2, 12, 0, 0, 0, time.UTC)},
		{name: "week", start: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), unit: "week", duration: 2, want: time.Date(2024, 2, 14, 12, 0, 0, 0, time.UTC)},
		{name: "month", start: time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), unit: "month", duration: 1, want: time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC)},
		{name: "year", start: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC), unit: "year", duration: 1, want: time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			got, err := calculateEndDate(test.start, test.unit, test.duration)

			// Then
			if err != nil {
				t.Fatalf("calculate end date: %v", err)
			}
			if !got.Equal(test.want) {
				t.Fatalf("end date = %s, want %s", got, test.want)
			}
		})
	}
}

func TestMembershipService_subscribe_concurrentlyLeavesOneActiveSubscription(t *testing.T) {
	// Given
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	seedPlan(t, client, "plan-1", "Starter", 25, 1, "month")
	service, err := NewMembershipService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new membership service: %v", err)
	}
	const attempts = 12
	errorsCh := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	waitGroup.Add(attempts)
	for range attempts {
		go func() {
			defer waitGroup.Done()
			_, subscribeErr := service.Subscribe(context.Background(), "user-1", "plan-1")
			errorsCh <- subscribeErr
		}()
	}
	waitGroup.Wait()
	close(errorsCh)
	for subscribeErr := range errorsCh {
		if subscribeErr != nil {
			t.Fatalf("concurrent subscribe: %v", subscribeErr)
		}
	}

	// Then
	active, err := client.Subscription.Query().Where(subscription.UserId("user-1"), subscription.IsActive(true)).Count(context.Background())
	if err != nil {
		t.Fatalf("count active subscriptions: %v", err)
	}
	if active != 1 {
		t.Fatalf("active subscriptions = %d, want 1", active)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.Credits != 100+attempts*25 {
		t.Fatalf("credits = %d, want %d", user.Credits, 100+attempts*25)
	}
	audits, err := client.AuditLog.Query().Count(context.Background())
	if err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if audits != attempts {
		t.Fatalf("audit events = %d, want %d", audits, attempts)
	}
}

func TestRedemptionService_redeem_concurrentlyClaimsCodeOnce(t *testing.T) {
	// Given
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	_, err := client.RedemptionCode.Create().SetID("code-1").SetCode("ONCEONLY").SetType("quota").SetCredits(50).SetCreatedBy("admin").SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background())
	if err != nil {
		t.Fatalf("seed redemption code: %v", err)
	}
	service, err := NewRedemptionService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new redemption service: %v", err)
	}
	const attempts = 12
	results := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	waitGroup.Add(attempts)
	for range attempts {
		go func() {
			defer waitGroup.Done()
			_, redeemErr := service.Redeem(context.Background(), "user-1", "onceonly")
			results <- redeemErr
		}()
	}
	waitGroup.Wait()
	close(results)
	successes := 0
	for redeemErr := range results {
		if redeemErr == nil {
			successes++
		}
	}

	// Then
	if successes != 1 {
		t.Fatalf("successful redemptions = %d, want 1", successes)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.Credits != 150 {
		t.Fatalf("credits = %d, want 150", user.Credits)
	}
	code, err := client.RedemptionCode.Get(context.Background(), "code-1")
	if err != nil {
		t.Fatalf("get redemption code: %v", err)
	}
	if !code.IsUsed || code.UsedBy != "user-1" {
		t.Fatalf("code usage = used:%v by:%q", code.IsUsed, code.UsedBy)
	}
	audits, err := client.AuditLog.Query().Count(context.Background())
	if err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if audits != 1 {
		t.Fatalf("audit events = %d, want 1", audits)
	}
}

func TestRedemptionService_subscriptionRedeemGrantsPlanCredits(t *testing.T) {
	// Given
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	seedPlan(t, client, "plan-1", "Starter", 25, 1, "month")
	_, err := client.RedemptionCode.Create().SetID("code-1").SetCode("SUBONLY").SetType("subscription").SetPlanId("plan-1").SetPlanName("Starter").SetCreatedBy("admin").SetCreatedAt(now).SetUpdatedAt(now).Save(context.Background())
	if err != nil {
		t.Fatalf("seed subscription code: %v", err)
	}
	service, err := NewRedemptionService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new redemption service: %v", err)
	}

	// When
	result, err := service.Redeem(context.Background(), "user-1", "subonly")

	// Then
	if err != nil {
		t.Fatalf("redeem subscription code: %v", err)
	}
	if result.Type != "subscription" {
		t.Fatalf("result type = %q, want subscription", result.Type)
	}
	user, err := client.User.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.Credits != 125 {
		t.Fatalf("credits = %d, want 125", user.Credits)
	}
	active, err := client.Subscription.Query().Where(subscription.UserId("user-1"), subscription.IsActive(true)).Count(context.Background())
	if err != nil {
		t.Fatalf("count active subscriptions: %v", err)
	}
	if active != 1 {
		t.Fatalf("active subscriptions = %d, want 1", active)
	}
}

func TestRegisterRoutes_subscribeReturnsSuccessEnvelope(t *testing.T) {
	// Given
	client, now := newMembershipClient(t)
	seedMembershipUser(t, client, "user-1", 100)
	seedPlan(t, client, "plan-1", "Starter", 25, 1, "month")
	plans, err := NewPlanService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new plan service: %v", err)
	}
	membership, err := NewMembershipService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new membership service: %v", err)
	}
	redemption, err := NewRedemptionService(client, fixedClock{now: now})
	if err != nil {
		t.Fatalf("new redemption service: %v", err)
	}
	handler, err := NewHandler(plans, membership, redemption)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/membership/subscribe", bytes.NewBufferString(`{"planId":"plan-1"}`))
	request = request.WithContext(auth.WithAuthContext(request.Context(), auth.AuthContext{User: auth.User{ID: "user-1", Role: "user"}}))
	recorder := httptest.NewRecorder()

	// When
	mux.ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Success bool              `json:"success"`
		Data    *ent.Subscription `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if !envelope.Success || envelope.Data == nil || envelope.Data.PlanId != "plan-1" {
		t.Fatalf("response envelope = %+v", envelope)
	}
}
