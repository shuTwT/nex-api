package payment

import (
	"context"
	"fmt"

	appRuntime "github.com/shuTwT/nex-api/internal/service/apierror"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/payment"
	pay "github.com/shuTwT/nex-api/internal/infra/pay"
)

func (s *Service) applyCallback(ctx context.Context, method pay.PaymentMethod, callback pay.ProviderCallback) (*ent.Payment, error) {
	if callback.OutTradeNo == "" {
		return nil, appRuntime.NewValidationError(appRuntime.FieldError{Field: "outTradeNo", Reason: "required"})
	}
	var updated *ent.Payment
	now := s.clock.Now().UTC()
	err := runTransaction(ctx, s.client, func(tx *ent.Tx) error {
		current, err := tx.Payment.Query().Where(payment.OutTradeNo(callback.OutTradeNo)).Only(ctx)
		if ent.IsNotFound(err) {
			return fmt.Errorf("payment %q: %w", callback.OutTradeNo, appRuntime.ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("load callback payment: %w", err)
		}
		if current.Method != string(method) {
			return fmt.Errorf("callback provider mismatch: %w", pay.ErrInvalidSignature)
		}
		if callback.HasAmount {
			if err := pay.ComparePaymentAmount(current.Amount, current.Currency, callback.Amount, callback.Currency); err != nil {
				return err
			}
		}
		if current.Status == string(pay.PaymentStatePaid) && callback.Status == pay.PaymentStatePaid {
			updated = current
			return nil
		}
		transition, err := pay.TransitionPayment(pay.PaymentState(current.Status), callback.Status, now)
		if err != nil {
			return err
		}
		builder := tx.Payment.UpdateOneID(current.ID).Where(payment.CallbackVersionEQ(current.CallbackVersion)).SetStatus(string(transition.Status)).SetCallbackVersion(current.CallbackVersion + 1).SetCallbackProcessedAt(now).SetUpdatedAt(now)
		switch transition.Status {
		case pay.PaymentStatePaid:
			builder.SetTransactionId(callback.TransactionID).SetPaidAt(callback.PaidAt).SetCallbackKey(callback.CallbackKey)
		case pay.PaymentStateCancelled:
			builder.SetCancelledAt(now).SetCallbackKey(callback.CallbackKey)
		default:
			builder.SetCallbackKey(callback.CallbackKey)
		}
		updated, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("update callback payment: %w", err)
		}
		if transition.Status == pay.PaymentStatePaid {
			if err := s.grantBusinessValue(ctx, tx, updated, now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
