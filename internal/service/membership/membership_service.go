package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/subscription"
	"github.com/shuTwT/nex-api/ent/subscriptionplan"
	appRuntime "github.com/shuTwT/nex-api/internal/service/apierror"
)

type MembershipService struct {
	client *ent.Client
	clock  Clock
}

func NewMembershipService(client *ent.Client, clocks ...Clock) (*MembershipService, error) {
	if client == nil {
		return nil, errors.New("membership: ent client is nil")
	}
	return &MembershipService{client: client, clock: selectClock(clocks)}, nil
}

func (s *MembershipService) ListPlans(ctx context.Context) ([]*ent.SubscriptionPlan, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	plans, err := s.client.SubscriptionPlan.Query().
		Where(subscriptionplan.IsActive(true)).
		Order(subscriptionplan.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active subscription plans: %w", err)
	}
	return plans, nil
}

func (s *MembershipService) Current(ctx context.Context, userID string) (*ent.Subscription, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	current, err := s.client.Subscription.Query().
		Where(
			subscription.UserId(userID),
			subscription.IsActive(true),
			subscription.EndDateGTE(now),
		).
		WithPlan().
		Order(subscription.ByCreatedAt(sql.OrderDesc())).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get current subscription: %w", err)
	}
	return current, nil
}

func (s *MembershipService) CurrentSubscription(ctx context.Context, userID string) (*ent.Subscription, error) {
	return s.Current(ctx, userID)
}

func (s *MembershipService) Subscribe(ctx context.Context, userID, planID string) (*ent.Subscription, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(planID) == "" {
		return nil, appRuntime.NewValidationError(appRuntime.FieldError{Field: "planId", Reason: "required"})
	}
	now := s.clock.Now().UTC()
	var created *ent.Subscription
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		plan, err := tx.SubscriptionPlan.Get(ctx, planID)
		if ent.IsNotFound(err) {
			return fmt.Errorf("subscription plan %q: %w", planID, appRuntime.ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("get subscription plan %q: %w", planID, err)
		}
		if !plan.IsActive {
			return appRuntime.NewError(appRuntime.KindInvalidInput, "plan_inactive", "subscription plan is inactive", appRuntime.ErrConflict)
		}
		created, err = activateSubscription(ctx, tx, userID, plan, plan.Price, now)
		if err != nil {
			return err
		}
		return writeAudit(ctx, tx, userID, "subscribe", "subscription", created.ID, map[string]string{"planId": plan.ID}, now)
	})
	if err != nil {
		return nil, err
	}
	return created.Unwrap(), nil
}

func (s *MembershipService) SubscribePlan(ctx context.Context, userID, planID string) (*ent.Subscription, error) {
	return s.Subscribe(ctx, userID, planID)
}

func calculateEndDate(start time.Time, unit string, duration int) (time.Time, error) {
	if duration < 1 {
		return time.Time{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityDuration", Reason: "must be positive"})
	}
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "day":
		return start.AddDate(0, 0, duration), nil
	case "week":
		return start.AddDate(0, 0, duration*7), nil
	case "month":
		return start.AddDate(0, duration, 0), nil
	case "year":
		return start.AddDate(duration, 0, 0), nil
	default:
		return time.Time{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "validityUnit", Reason: "must be day, week, month, or year"})
	}
}

func createSubscription(ctx context.Context, tx *ent.Tx, userID string, plan *ent.SubscriptionPlan, price float64, start time.Time) (*ent.Subscription, error) {
	end, err := calculateEndDate(start, plan.ValidityUnit, plan.ValidityDuration)
	if err != nil {
		return nil, err
	}
	created, err := tx.Subscription.Create().
		SetUserId(userID).
		SetPlanId(plan.ID).
		SetPlanName(plan.Title).
		SetCredits(plan.TotalCredits).
		SetPrice(price).
		SetStartDate(start).
		SetEndDate(end).
		SetIsActive(true).
		SetCreatedAt(start).
		SetUpdatedAt(start).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("create subscription: %w", appRuntime.ErrConflict)
		}
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	return created, nil
}

func activateSubscription(ctx context.Context, tx *ent.Tx, userID string, plan *ent.SubscriptionPlan, price float64, now time.Time) (*ent.Subscription, error) {
	if err := updateUserCredits(ctx, tx, userID, 0, now); err != nil {
		return nil, err
	}
	if _, err := tx.Subscription.Update().
		Where(subscription.UserId(userID), subscription.IsActive(true)).
		SetIsActive(false).
		SetUpdatedAt(now).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("deactivate current subscriptions: %w", err)
	}
	created, err := createSubscription(ctx, tx, userID, plan, price, now)
	if err != nil {
		return nil, err
	}
	if err := updateUserCredits(ctx, tx, userID, plan.TotalCredits, now); err != nil {
		return nil, err
	}
	return created, nil
}

func updateUserCredits(ctx context.Context, tx *ent.Tx, userID string, credits int, now time.Time) error {
	_, err := tx.User.UpdateOneID(userID).
		AddCredits(credits).
		SetUpdatedAt(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return fmt.Errorf("user %q: %w", userID, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("update user credits: %w", err)
	}
	return nil
}
