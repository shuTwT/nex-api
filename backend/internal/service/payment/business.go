package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/service/apierror"
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/subscription"
)

type paymentMetadata struct {
	Type        string  `json:"type"`
	PlanID      string  `json:"planId"`
	Credits     int     `json:"credits"`
	CreditPrice float64 `json:"creditPrice"`
}

func (s *Service) grantBusinessValue(ctx context.Context, tx *ent.Tx, record *ent.Payment, now time.Time) error {
	var metadata paymentMetadata
	if err := json.Unmarshal([]byte(record.Metadata), &metadata); err != nil {
		return fmt.Errorf("decode payment metadata: %w", err)
	}
	switch metadata.Type {
	case "recharge":
		if metadata.Credits <= 0 {
			return appRuntime.NewValidationError(appRuntime.FieldError{Field: "credits", Reason: "must be positive"})
		}
		return addCredits(ctx, tx, record.UserId, metadata.Credits, now)
	case "subscription":
		return activatePaidSubscription(ctx, tx, record, metadata.PlanID, now)
	default:
		return nil
	}
}

func activatePaidSubscription(ctx context.Context, tx *ent.Tx, record *ent.Payment, planID string, now time.Time) error {
	if planID == "" {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "planId", Reason: "required"})
	}
	plan, err := tx.SubscriptionPlan.Get(ctx, planID)
	if ent.IsNotFound(err) {
		return fmt.Errorf("subscription plan %q: %w", planID, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("load paid subscription plan: %w", err)
	}
	if _, err := tx.Subscription.Update().Where(subscription.UserId(record.UserId), subscription.IsActive(true)).SetIsActive(false).SetUpdatedAt(now).Save(ctx); err != nil {
		return fmt.Errorf("deactivate subscription: %w", err)
	}
	endDate, err := subscriptionEndDate(now, plan.ValidityUnit, plan.ValidityDuration)
	if err != nil {
		return err
	}
	if _, err := tx.Subscription.Create().SetUserId(record.UserId).SetPlanId(plan.ID).SetPlanName(plan.Title).SetCredits(plan.TotalCredits).SetPrice(record.Amount).SetStartDate(now).SetEndDate(endDate).SetIsActive(true).SetPaymentId(record.ID).SetCreatedAt(now).SetUpdatedAt(now).Save(ctx); err != nil {
		return fmt.Errorf("create paid subscription: %w", err)
	}
	return addCredits(ctx, tx, record.UserId, plan.TotalCredits, now)
}

func addCredits(ctx context.Context, tx *ent.Tx, userID string, credits int, now time.Time) error {
	if _, err := tx.User.UpdateOneID(userID).AddCredits(credits).SetUpdatedAt(now).Save(ctx); err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("user %q: %w", userID, appRuntime.ErrNotFound)
		}
		return fmt.Errorf("grant credits: %w", err)
	}
	return nil
}

func subscriptionEndDate(start time.Time, unit string, duration int) (time.Time, error) {
	if duration < 1 {
		return time.Time{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityDuration", Reason: "must be positive"})
	}
	switch strings.ToLower(unit) {
	case "day":
		return start.AddDate(0, 0, duration), nil
	case "week":
		return start.AddDate(0, 0, duration*7), nil
	case "month":
		return start.AddDate(0, duration, 0), nil
	case "year":
		return start.AddDate(duration, 0, 0), nil
	default:
		return time.Time{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityUnit", Reason: "unsupported"})
	}
}

func parseSettingFloat(values map[string]string, key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(values[key], 64)
	if err != nil {
		return fallback
	}
	return value
}
