package membership

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
	servicemembership "github.com/shuTwT/nex-api/backend/internal/service/membership"
)

func TestRegisterRoutes_subscribeReturnsSuccessEnvelope(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/membership.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	if _, err := client.User.Create().SetID("user-1").SetEmail("user@example.com").SetUsername("user").SetPassword("redacted").SetRole("user").SetCredits(100).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.SubscriptionPlan.Create().SetID("plan-1").SetTitle("Starter").SetPrice(10).SetTotalCredits(25).SetSortOrder(0).SetValidityDuration(1).SetValidityUnit("month").SetCreditResetCycle("month").SetIsActive(true).Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	plans, err := servicemembership.NewPlanService(client)
	if err != nil {
		t.Fatal(err)
	}
	membership, err := servicemembership.NewMembershipService(client)
	if err != nil {
		t.Fatal(err)
	}
	redemption, err := servicemembership.NewRedemptionService(client)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(plans, membership, redemption)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/membership/subscribe", strings.NewReader(`{"planId":"plan-1"}`))
	request = request.WithContext(serviceauth.WithAuthContext(request.Context(), serviceauth.AuthContext{User: serviceauth.User{ID: "user-1", Role: "user"}}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			PlanID string `json:"planId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.Success || envelope.Data.PlanID != "plan-1" {
		t.Fatalf("response = %s", recorder.Body.String())
	}
}
