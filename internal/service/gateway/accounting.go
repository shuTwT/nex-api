package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/apiusage"
	"github.com/shuTwT/nex-api/ent/user"
)

var (
	ErrInsufficientCredits = errors.New("gateway: insufficient credits")
	ErrReservationState    = errors.New("gateway: invalid reservation state")
)

const (
	reservationStatus = "reserved"
	successStatus     = "success"
	failedStatus      = "failed"
)

type CreditRequest struct {
	UserID  string
	APIID   string
	Credits int
}

type CreditReservation struct {
	ID      string
	UserID  string
	APIID   string
	Credits int
}

type Accountant interface {
	Reserve(context.Context, CreditRequest) (CreditReservation, error)
	Finalize(context.Context, CreditReservation) error
	Refund(context.Context, CreditReservation) error
}

type EntAccountant struct{ db *ent.Client }

func NewEntAccountant(db *ent.Client) *EntAccountant { return &EntAccountant{db: db} }

func (a *EntAccountant) Reserve(ctx context.Context, request CreditRequest) (CreditReservation, error) {
	if request.Credits < 0 {
		return CreditReservation{}, fmt.Errorf("negative pricing: %w", ErrInsufficientCredits)
	}
	tx, err := a.db.Tx(ctx)
	if err != nil {
		return CreditReservation{}, fmt.Errorf("begin credit reservation: %w", err)
	}
	updated, err := tx.User.Update().Where(user.IDEQ(request.UserID), user.CreditsGTE(request.Credits)).AddCredits(-request.Credits).Save(ctx)
	if err != nil {
		return CreditReservation{}, rollback(tx, fmt.Errorf("reserve credits: %w", err))
	}
	if updated != 1 {
		return CreditReservation{}, rollback(tx, ErrInsufficientCredits)
	}
	usage, err := tx.ApiUsage.Create().SetUserId(request.UserID).SetApiId(request.APIID).SetCredits(request.Credits).SetStatus(reservationStatus).Save(ctx)
	if err != nil {
		return CreditReservation{}, rollback(tx, fmt.Errorf("create reservation: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return CreditReservation{}, fmt.Errorf("commit credit reservation: %w", err)
	}
	return CreditReservation{ID: usage.ID, UserID: request.UserID, APIID: request.APIID, Credits: request.Credits}, nil
}

func (a *EntAccountant) Finalize(ctx context.Context, reservation CreditReservation) error {
	tx, err := a.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin credit finalization: %w", err)
	}
	updated, err := tx.ApiUsage.Update().Where(apiusage.IDEQ(reservation.ID), apiusage.StatusEQ(reservationStatus)).SetStatus(successStatus).Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("finalize usage: %w", err))
	}
	if updated != 1 {
		return rollback(tx, ErrReservationState)
	}
	if _, err := tx.Api.UpdateOneID(reservation.APIID).AddCallCount(1).Save(ctx); err != nil {
		return rollback(tx, fmt.Errorf("increment API call count: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit credit finalization: %w", err)
	}
	return nil
}

func (a *EntAccountant) Refund(ctx context.Context, reservation CreditReservation) error {
	tx, err := a.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin credit refund: %w", err)
	}
	updated, err := tx.ApiUsage.Update().Where(apiusage.IDEQ(reservation.ID), apiusage.StatusEQ(reservationStatus)).SetStatus(failedStatus).Save(ctx)
	if err != nil {
		return rollback(tx, fmt.Errorf("mark usage failed: %w", err))
	}
	if updated != 1 {
		return rollback(tx, ErrReservationState)
	}
	if _, err := tx.User.UpdateOneID(reservation.UserID).AddCredits(reservation.Credits).Save(ctx); err != nil {
		return rollback(tx, fmt.Errorf("refund credits: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit credit refund: %w", err)
	}
	return nil
}

func rollback(tx *ent.Tx, err error) error {
	return errors.Join(err, tx.Rollback())
}
