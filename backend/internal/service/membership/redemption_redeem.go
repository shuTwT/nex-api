package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/ent/redemptioncode"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/service/apierror"
)

func (s *RedemptionService) Lookup(ctx context.Context, codeInput string) (RedemptionLookup, error) {
	if err := validateContext(ctx); err != nil {
		return RedemptionLookup{}, err
	}
	code := normalizeCode(codeInput)
	if code == "" {
		return RedemptionLookup{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "code", Reason: "required"})
	}
	item, err := s.client.RedemptionCode.Query().Where(redemptioncode.Code(code)).Only(ctx)
	if ent.IsNotFound(err) {
		return RedemptionLookup{}, fmt.Errorf("redemption code %q: %w", code, appRuntime.ErrNotFound)
	}
	if err != nil {
		return RedemptionLookup{}, fmt.Errorf("lookup redemption code: %w", err)
	}
	if item.IsUsed {
		return RedemptionLookup{}, appRuntime.NewError(appRuntime.KindInvalidInput, "code_used", "redemption code has already been used", appRuntime.ErrConflict)
	}
	if item.ExpiresAt.Before(s.clock.Now().UTC()) && !item.ExpiresAt.IsZero() {
		return RedemptionLookup{}, appRuntime.NewError(appRuntime.KindInvalidInput, "code_expired", "redemption code has expired", appRuntime.ErrConflict)
	}
	lookup := RedemptionLookup{Type: item.Type}
	if item.PlanName != "" {
		lookup.PlanName = &item.PlanName
	}
	if item.Type == "quota" {
		credits := item.Credits
		lookup.Credits = &credits
	}
	return lookup, nil
}

func (s *RedemptionService) Redeem(ctx context.Context, userID, codeInput string) (RedemptionResult, error) {
	if err := validateContext(ctx); err != nil {
		return RedemptionResult{}, err
	}
	if err := validateUserID(userID); err != nil {
		return RedemptionResult{}, err
	}
	code := normalizeCode(codeInput)
	if code == "" {
		return RedemptionResult{}, appRuntime.NewValidationError(appRuntime.FieldError{Field: "code", Reason: "required"})
	}
	now := s.clock.Now().UTC()
	var result RedemptionResult
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		claimed, err := tx.RedemptionCode.Update().
			Where(
				redemptioncode.Code(code),
				redemptioncode.IsUsed(false),
				redemptioncode.Or(redemptioncode.ExpiresAtIsNil(), redemptioncode.ExpiresAtGTE(now)),
			).
			SetIsUsed(true).
			SetUsedBy(userID).
			SetUsedAt(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("claim redemption code: %w", err)
		}
		if claimed != 1 {
			return classifyUnavailableCode(ctx, tx, code, now)
		}
		item, err := tx.RedemptionCode.Query().Where(redemptioncode.Code(code)).Only(ctx)
		if err != nil {
			return fmt.Errorf("load claimed redemption code: %w", err)
		}
		switch item.Type {
		case "quota":
			if err := updateUserCredits(ctx, tx, userID, item.Credits, now); err != nil {
				return err
			}
			result = RedemptionResult{Message: fmt.Sprintf("兑换成功！获得 %d 额度", item.Credits), Type: item.Type, Credits: item.Credits}
		case "subscription":
			plan, planErr := tx.SubscriptionPlan.Get(ctx, item.PlanId)
			if ent.IsNotFound(planErr) {
				return fmt.Errorf("redemption plan %q: %w", item.PlanId, appRuntime.ErrNotFound)
			}
			if planErr != nil {
				return fmt.Errorf("get redemption plan: %w", planErr)
			}
			if _, err := activateSubscription(ctx, tx, userID, plan, 0, now); err != nil {
				return err
			}
			result = RedemptionResult{Message: fmt.Sprintf("兑换成功！获得 %s 订阅", item.PlanName), Type: item.Type}
		default:
			return appRuntime.NewError(appRuntime.KindInvalidInput, "invalid_code_type", "redemption code type is invalid", appRuntime.ErrValidation)
		}
		return writeAudit(ctx, tx, userID, "redeem", "redemption_code", item.ID, map[string]string{"code": code}, now)
	})
	if err != nil {
		return RedemptionResult{}, err
	}
	return result, nil
}

func (s *RedemptionService) RedeemCode(ctx context.Context, userID, codeInput string) (RedemptionResult, error) {
	return s.Redeem(ctx, userID, codeInput)
}

func classifyUnavailableCode(ctx context.Context, tx *ent.Tx, code string, now time.Time) error {
	item, err := tx.RedemptionCode.Query().Where(redemptioncode.Code(code)).Only(ctx)
	if ent.IsNotFound(err) {
		return fmt.Errorf("redemption code %q: %w", code, appRuntime.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("inspect redemption code: %w", err)
	}
	if item.IsUsed {
		return appRuntime.NewError(appRuntime.KindInvalidInput, "code_used", "redemption code has already been used", appRuntime.ErrConflict)
	}
	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(now) {
		return appRuntime.NewError(appRuntime.KindInvalidInput, "code_expired", "redemption code has expired", appRuntime.ErrConflict)
	}
	return appRuntime.NewError(appRuntime.KindInvalidInput, "code_unavailable", "redemption code cannot be redeemed", appRuntime.ErrConflict)
}
