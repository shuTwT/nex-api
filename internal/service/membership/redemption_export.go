package membership

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/redemptioncode"
	"github.com/shuTwT/nex-api/ent/subscriptionplan"
)

func (s *RedemptionService) Export(ctx context.Context, ids []string) (string, error) {
	if err := validateContext(ctx); err != nil {
		return "", err
	}
	query := s.client.RedemptionCode.Query().Order(redemptioncode.ByCreatedAt(sql.OrderDesc()))
	ids = uniqueNonEmpty(ids)
	if len(ids) > 0 {
		query.Where(redemptioncode.IDIn(ids...))
	}
	items, err := query.All(ctx)
	if err != nil {
		return "", fmt.Errorf("export redemption codes: %w", err)
	}
	var builder strings.Builder
	builder.WriteString("\uFEFF兑换码,类型,订阅计划,额度,过期时间,状态,创建时间\n")
	writer := csv.NewWriter(&builder)
	for _, item := range items {
		credits := "-"
		planName := "-"
		if item.Type == "quota" {
			credits = fmt.Sprintf("%d", item.Credits)
		}
		if item.Type == "subscription" && item.PlanName != "" {
			planName = item.PlanName
		}
		expires := "永久"
		if !item.ExpiresAt.IsZero() {
			expires = item.ExpiresAt.UTC().Format(time.RFC3339)
		}
		if err := writer.Write([]string{item.Code, item.Type, planName, credits, expires, usedLabel(item.IsUsed), item.CreatedAt.UTC().Format(time.RFC3339)}); err != nil {
			return "", fmt.Errorf("write redemption CSV: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush redemption CSV: %w", err)
	}
	return builder.String(), nil
}

// ListPlans returns the active subscription plans for the redemption form.
func (s *RedemptionService) ListPlans(ctx context.Context) ([]*ent.SubscriptionPlan, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if s.client == nil {
		return nil, fmt.Errorf("membership: ent client is nil")
	}
	plans, err := s.client.SubscriptionPlan.Query().Where(subscriptionplan.IsActive(true)).Order(subscriptionplan.BySortOrder()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list redemption plans: %w", err)
	}
	return plans, nil
}
