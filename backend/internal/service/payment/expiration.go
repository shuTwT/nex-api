package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/shuTwT/nex-api/backend/ent"
	dbPayment "github.com/shuTwT/nex-api/backend/ent/payment"
	pay "github.com/shuTwT/nex-api/backend/internal/infra/pay"
)

// ExpirePendingPayments closes due provider orders and transitions their local
// state from pending to expired. A provider/query failure leaves the order in
// pending so the next scheduled run can retry safely.
func (s *Service) ExpirePendingPayments(ctx context.Context) (int, error) {
	now := s.clock.Now().UTC()
	items, err := s.client.Payment.Query().Where(
		dbPayment.StatusEQ(string(pay.PaymentStatePending)),
		dbPayment.ExpiredAtLTE(now),
	).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list expired pending payments: %w", err)
	}

	expired := 0
	var runErrors []error
	for _, item := range items {
		changed, processErr := s.expirePendingPayment(ctx, item)
		if processErr != nil {
			runErrors = append(runErrors, fmt.Errorf("expire payment %s: %w", item.OutTradeNo, processErr))
			continue
		}
		if changed {
			expired++
		}
	}
	return expired, errors.Join(runErrors...)
}

func (s *Service) expirePendingPayment(ctx context.Context, item *ent.Payment) (bool, error) {
	provider, _, err := s.resolveProvider(ctx, pay.PaymentMethod(item.Method))
	if err != nil {
		return false, err
	}
	if item.Method != string(pay.PaymentMethodMock) {
		status, queryErr := provider.Query(ctx, item.OutTradeNo)
		if queryErr != nil {
			return false, fmt.Errorf("query provider before expiration: %w", queryErr)
		}
		if status.Status != pay.PaymentStatePending {
			callback := pay.ProviderCallback{
				OutTradeNo: item.OutTradeNo, TransactionID: status.TransactionID,
				Status: status.Status, PaidAt: status.PaidAt,
				CallbackKey: string(item.Method) + ":expiry-query:" + item.OutTradeNo,
			}
			if status.Status == pay.PaymentStatePaid {
				callback.Amount = status.Amount
				callback.Currency = status.Currency
				callback.HasAmount = true
			}
			_, err := s.applyCallback(ctx, pay.PaymentMethod(item.Method), callback)
			return false, err
		}
	}
	if err := provider.Cancel(ctx, item.OutTradeNo); err != nil {
		return false, fmt.Errorf("close provider order: %w", err)
	}

	now := s.clock.Now().UTC()
	updated, err := s.client.Payment.UpdateOneID(item.ID).Where(
		dbPayment.StatusEQ(string(pay.PaymentStatePending)),
		dbPayment.CallbackVersionEQ(item.CallbackVersion),
		dbPayment.ExpiredAtLTE(now),
	).SetStatus(string(pay.PaymentStateExpired)).
		SetCallbackVersion(item.CallbackVersion + 1).
		SetUpdatedAt(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		// A callback or another scheduler instance won the race.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("persist expired status: %w", err)
	}
	return updated.Status == string(pay.PaymentStateExpired), nil
}
