package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/redemptioncode"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/runtime"
)

const redemptionAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type RedemptionCreateInput struct {
	Type      string
	Count     int
	PlanID    string
	Credits   int
	ExpiresAt *time.Time
}

type RedemptionListFilter struct {
	Search string
	Type   string
	IsUsed *bool
	Page   int
	Limit  int
}

type RedemptionPage struct {
	Items      []*ent.RedemptionCode
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type RedemptionBatchResult struct {
	Count   int    `json:"count"`
	BatchID string `json:"batchId"`
}

type RedemptionLookup struct {
	Type     string  `json:"type"`
	PlanName *string `json:"planName"`
	Credits  *int    `json:"credits"`
}

type RedemptionResult struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Credits int    `json:"credits,omitempty"`
}

type RedemptionService struct {
	client *ent.Client
	clock  Clock
}

func NewRedemptionService(client *ent.Client, clocks ...Clock) (*RedemptionService, error) {
	if client == nil {
		return nil, errors.New("membership: ent client is nil")
	}
	return &RedemptionService{client: client, clock: selectClock(clocks)}, nil
}

func (s *RedemptionService) List(ctx context.Context, filter RedemptionListFilter) (RedemptionPage, error) {
	if err := validateContext(ctx); err != nil {
		return RedemptionPage{}, err
	}
	filter = normalizeRedemptionFilter(filter)
	query := s.client.RedemptionCode.Query()
	if search := strings.TrimSpace(filter.Search); search != "" {
		query.Where(redemptioncode.CodeContains(search))
	}
	if codeType := strings.TrimSpace(filter.Type); codeType != "" && codeType != "all" {
		query.Where(redemptioncode.Type(codeType))
	}
	if filter.IsUsed != nil {
		query.Where(redemptioncode.IsUsed(*filter.IsUsed))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return RedemptionPage{}, fmt.Errorf("count redemption codes: %w", err)
	}
	items, err := query.Order(redemptioncode.ByCreatedAt(sql.OrderDesc())).
		Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).All(ctx)
	if err != nil {
		return RedemptionPage{}, fmt.Errorf("list redemption codes: %w", err)
	}
	return RedemptionPage{Items: items, Total: total, Page: filter.Page, PageSize: filter.Limit, TotalPages: pageCount(total, filter.Limit)}, nil
}

func (s *RedemptionService) ListCodes(ctx context.Context, filter RedemptionListFilter) (RedemptionPage, error) {
	return s.List(ctx, filter)
}

func (s *RedemptionService) CreateBatch(ctx context.Context, actorID string, input RedemptionCreateInput) (RedemptionBatchResult, error) {
	if err := validateContext(ctx); err != nil {
		return RedemptionBatchResult{}, err
	}
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	if input.Count < 1 || input.Count > 1000 {
		return RedemptionBatchResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "count", Reason: "must be between 1 and 1000"})
	}
	if input.Type != "quota" && input.Type != "subscription" {
		return RedemptionBatchResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "type", Reason: "must be quota or subscription"})
	}
	if input.Type == "quota" && input.Credits < 1 {
		return RedemptionBatchResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "credits", Reason: "must be positive"})
	}
	if input.Type == "subscription" && strings.TrimSpace(input.PlanID) == "" {
		return RedemptionBatchResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "planId", Reason: "required"})
	}
	if input.ExpiresAt != nil {
		expires := input.ExpiresAt.UTC()
		input.ExpiresAt = &expires
	}
	codes, err := generateCodes(input.Count)
	if err != nil {
		return RedemptionBatchResult{}, err
	}
	now := s.clock.Now().UTC()
	result := RedemptionBatchResult{Count: input.Count, BatchID: uuid.NewString()}
	err = runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		var planName string
		if input.Type == "subscription" {
			plan, planErr := tx.SubscriptionPlan.Get(ctx, input.PlanID)
			if ent.IsNotFound(planErr) {
				return fmt.Errorf("subscription plan %q: %w", input.PlanID, appRuntime.ErrNotFound)
			}
			if planErr != nil {
				return fmt.Errorf("get redemption plan: %w", planErr)
			}
			planName = plan.Title
		}
		builders := make([]*ent.RedemptionCodeCreate, 0, len(codes))
		for _, code := range codes {
			builder := tx.RedemptionCode.Create().
				SetCode(code).
				SetType(input.Type).
				SetCreatedBy(actorID).
				SetBatchId(result.BatchID).
				SetCreatedAt(now).
				SetUpdatedAt(now)
			if input.Type == "quota" {
				builder.SetCredits(input.Credits)
			} else {
				builder.SetPlanId(input.PlanID).SetPlanName(planName)
			}
			builder.SetNillableExpiresAt(input.ExpiresAt)
			builders = append(builders, builder)
		}
		if _, err := tx.RedemptionCode.CreateBulk(builders...).Save(ctx); err != nil {
			if ent.IsConstraintError(err) {
				return fmt.Errorf("create redemption codes: %w", appRuntime.ErrConflict)
			}
			return fmt.Errorf("create redemption codes: %w", err)
		}
		return writeAudit(ctx, tx, actorID, "create", "redemption_code", result.BatchID, input, now)
	})
	if err != nil {
		return RedemptionBatchResult{}, err
	}
	return result, nil
}

func (s *RedemptionService) CreateCodes(ctx context.Context, actorID string, input RedemptionCreateInput) (RedemptionBatchResult, error) {
	return s.CreateBatch(ctx, actorID, input)
}

func (s *RedemptionService) Delete(ctx context.Context, actorID, id string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "id", Reason: "required"})
	}
	now := s.clock.Now().UTC()
	return runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		deleted, err := tx.RedemptionCode.Delete().Where(redemptioncode.ID(id), redemptioncode.IsUsed(false)).Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete redemption code: %w", err)
		}
		if deleted != 1 {
			return appRuntime.NewAPIError(400, "code_unavailable", "redemption code is missing or already used", appRuntime.ErrConflict)
		}
		return writeAudit(ctx, tx, actorID, "delete", "redemption_code", id, nil, now)
	})
}

func (s *RedemptionService) DeleteBatch(ctx context.Context, actorID, batchID string) (int, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}
	if strings.TrimSpace(batchID) == "" {
		return 0, appRuntime.NewValidationError(appRuntime.FieldError{Field: "batchId", Reason: "required"})
	}
	now := s.clock.Now().UTC()
	deleted := 0
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		var err error
		deleted, err = tx.RedemptionCode.Delete().Where(redemptioncode.BatchId(batchID), redemptioncode.IsUsed(false)).Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete redemption code batch: %w", err)
		}
		if deleted == 0 {
			return fmt.Errorf("redemption batch %q: %w", batchID, appRuntime.ErrNotFound)
		}
		used, err := tx.RedemptionCode.Query().Where(redemptioncode.BatchId(batchID), redemptioncode.IsUsed(true)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("check redemption code batch: %w", err)
		}
		if used {
			return appRuntime.NewAPIError(400, "batch_contains_used_codes", "batch contains used redemption codes", appRuntime.ErrConflict)
		}
		return writeAudit(ctx, tx, actorID, "delete", "redemption_code_batch", batchID, nil, now)
	})
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func (s *RedemptionService) DeleteBatchByIDs(ctx context.Context, actorID string, ids []string) (int, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}
	ids = uniqueNonEmpty(ids)
	if len(ids) == 0 {
		return 0, appRuntime.NewValidationError(appRuntime.FieldError{Field: "ids", Reason: "must not be empty"})
	}
	now := s.clock.Now().UTC()
	deleted := 0
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		var err error
		deleted, err = tx.RedemptionCode.Delete().Where(redemptioncode.IDIn(ids...), redemptioncode.IsUsed(false)).Exec(ctx)
		if err != nil {
			return fmt.Errorf("delete redemption codes: %w", err)
		}
		if deleted != len(ids) {
			return appRuntime.NewAPIError(400, "codes_unavailable", "one or more redemption codes are missing or used", appRuntime.ErrConflict)
		}
		return writeAudit(ctx, tx, actorID, "delete", "redemption_code_batch", strings.Join(ids, ","), nil, now)
	})
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
