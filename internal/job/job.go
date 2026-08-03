// Package job registers the built-in scheduled tasks on the schedule
// manager. Tasks themselves call into the service layer.
package job

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shuTwT/nex-api/internal/infra/schedule"
	servicepayment "github.com/shuTwT/nex-api/internal/service/payment"
)

// Dependencies carries the services the built-in jobs invoke.
type Dependencies struct {
	StatsSync func(context.Context) error
	Payments  *servicepayment.Service
	Logger    *slog.Logger
}

// RegisterAll registers the two built-in jobs: stats synchronization and
// pending payment expiration.
func RegisterAll(manager *schedule.ScheduleManager, deps Dependencies) error {
	if manager == nil {
		return errors.New("job: schedule manager is required")
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	if err := manager.RegisterTask(schedule.TaskKeyStatsSync, "将 Redis 中的 API/MCP 调用统计同步到数据库", deps.StatsSync); err != nil {
		return err
	}
	if err := manager.RegisterTask(schedule.TaskKeyExpirePayments, "关闭并过期超过支付时限的 pending 订单", func(ctx context.Context) error {
		if deps.Payments == nil {
			return errors.New("job: payment service is required for payment expiration")
		}
		count, expireErr := deps.Payments.ExpirePendingPayments(ctx)
		if count > 0 {
			logger.Info("expired pending payment orders", slog.Int("count", count))
		}
		return expireErr
	}); err != nil {
		return err
	}
	return nil
}
