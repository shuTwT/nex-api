package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/subscriptionplan"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

type PlanCreateInput struct {
	Title            string
	Price            float64
	TotalCredits     int
	SortOrder        int
	ValidityDuration int
	ValidityUnit     string
	CreditResetCycle string
	IsActive         bool
}

type PlanUpdateInput struct {
	Title            *string
	Price            *float64
	TotalCredits     *int
	SortOrder        *int
	ValidityDuration *int
	ValidityUnit     *string
	CreditResetCycle *string
	IsActive         *bool
}

type PlanListFilter struct {
	Search   string
	IsActive *bool
	Page     int
	Limit    int
}

type PlanPage struct {
	Items      []*ent.SubscriptionPlan
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type PlanService struct {
	client *ent.Client
	clock  Clock
}

func NewPlanService(client *ent.Client, clocks ...Clock) (*PlanService, error) {
	if client == nil {
		return nil, errors.New("membership: ent client is nil")
	}
	return &PlanService{client: client, clock: selectClock(clocks)}, nil
}

func (s *PlanService) List(ctx context.Context, filter PlanListFilter) (PlanPage, error) {
	if err := validateContext(ctx); err != nil {
		return PlanPage{}, err
	}
	filter = normalizePageFilter(filter)
	query := s.client.SubscriptionPlan.Query()
	if search := strings.TrimSpace(filter.Search); search != "" {
		query.Where(subscriptionplan.TitleContains(search))
	}
	if filter.IsActive != nil {
		query.Where(subscriptionplan.IsActive(*filter.IsActive))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return PlanPage{}, fmt.Errorf("list subscription plans count: %w", err)
	}
	items, err := query.Order(subscriptionplan.BySortOrder()).Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).All(ctx)
	if err != nil {
		return PlanPage{}, fmt.Errorf("list subscription plans: %w", err)
	}
	return PlanPage{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.Limit,
		TotalPages: pageCount(total, filter.Limit),
	}, nil
}

func (s *PlanService) ListPlans(ctx context.Context, filter PlanListFilter) (PlanPage, error) {
	return s.List(ctx, filter)
}

func (s *PlanService) Get(ctx context.Context, id string) (*ent.SubscriptionPlan, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, appRuntime.NewValidationError(appRuntime.FieldError{Field: "id", Reason: "required"})
	}
	plan, err := s.client.SubscriptionPlan.Get(ctx, id)
	if ent.IsNotFound(err) {
		return nil, fmt.Errorf("subscription plan %q: %w", id, appRuntime.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription plan %q: %w", id, err)
	}
	return plan, nil
}

func (s *PlanService) GetPlan(ctx context.Context, id string) (*ent.SubscriptionPlan, error) {
	return s.Get(ctx, id)
}

func (s *PlanService) Create(ctx context.Context, actorID string, input PlanCreateInput) (*ent.SubscriptionPlan, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	input, err := normalizeCreateInput(input)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	var plan *ent.SubscriptionPlan
	err = runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		var saveErr error
		plan, saveErr = tx.SubscriptionPlan.Create().
			SetTitle(input.Title).
			SetPrice(input.Price).
			SetTotalCredits(input.TotalCredits).
			SetSortOrder(input.SortOrder).
			SetValidityDuration(input.ValidityDuration).
			SetValidityUnit(input.ValidityUnit).
			SetCreditResetCycle(input.CreditResetCycle).
			SetIsActive(input.IsActive).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if saveErr != nil {
			return wrapPlanWriteError("create subscription plan", saveErr)
		}
		return writeAudit(ctx, tx, actorID, "create", "subscription_plan", plan.ID, input, now)
	})
	if err != nil {
		return nil, err
	}
	return plan.Unwrap(), nil
}

func (s *PlanService) CreatePlan(ctx context.Context, actorID string, input PlanCreateInput) (*ent.SubscriptionPlan, error) {
	return s.Create(ctx, actorID, input)
}

func (s *PlanService) Update(ctx context.Context, actorID, id string, input PlanUpdateInput) (*ent.SubscriptionPlan, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, appRuntime.NewValidationError(appRuntime.FieldError{Field: "id", Reason: "required"})
	}
	if err := validateUpdateInput(&input); err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	var plan *ent.SubscriptionPlan
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		builder := tx.SubscriptionPlan.UpdateOneID(id).SetUpdatedAt(now)
		if input.Title != nil {
			builder.SetTitle(strings.TrimSpace(*input.Title))
		}
		if input.Price != nil {
			builder.SetPrice(*input.Price)
		}
		if input.TotalCredits != nil {
			builder.SetTotalCredits(*input.TotalCredits)
		}
		if input.SortOrder != nil {
			builder.SetSortOrder(*input.SortOrder)
		}
		if input.ValidityDuration != nil {
			builder.SetValidityDuration(*input.ValidityDuration)
		}
		if input.ValidityUnit != nil {
			builder.SetValidityUnit(strings.ToLower(strings.TrimSpace(*input.ValidityUnit)))
		}
		if input.CreditResetCycle != nil {
			builder.SetCreditResetCycle(strings.ToLower(strings.TrimSpace(*input.CreditResetCycle)))
		}
		if input.IsActive != nil {
			builder.SetIsActive(*input.IsActive)
		}
		var saveErr error
		plan, saveErr = builder.Save(ctx)
		if saveErr != nil {
			return wrapPlanWriteError("update subscription plan", saveErr)
		}
		return writeAudit(ctx, tx, actorID, "update", "subscription_plan", id, input, now)
	})
	if err != nil {
		return nil, err
	}
	return plan.Unwrap(), nil
}

func (s *PlanService) UpdatePlan(ctx context.Context, actorID, id string, input PlanUpdateInput) (*ent.SubscriptionPlan, error) {
	return s.Update(ctx, actorID, id, input)
}

func (s *PlanService) Delete(ctx context.Context, actorID, id string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "id", Reason: "required"})
	}
	now := s.clock.Now().UTC()
	return runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		deleted, err := tx.SubscriptionPlan.Delete().Where(subscriptionplan.ID(id)).Exec(ctx)
		if err != nil {
			return wrapPlanWriteError("delete subscription plan", err)
		}
		if deleted != 1 {
			return fmt.Errorf("subscription plan %q: %w", id, appRuntime.ErrNotFound)
		}
		return writeAudit(ctx, tx, actorID, "delete", "subscription_plan", id, nil, now)
	})
}

func (s *PlanService) DeletePlan(ctx context.Context, actorID, id string) error {
	return s.Delete(ctx, actorID, id)
}
