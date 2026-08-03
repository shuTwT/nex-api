package membership

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/ent"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/service/apierror"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func selectClock(clocks []Clock) Clock {
	if len(clocks) > 0 && clocks[0] != nil {
		return clocks[0]
	}
	return realClock{}
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("membership: context is nil")
	}
	return nil
}

func validateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "userId", Reason: "required"})
	}
	return nil
}

func writeAudit(ctx context.Context, tx *ent.Tx, userID, action, resource, resourceID string, details interface{}, now time.Time) error {
	detailBytes, err := json.Marshal(struct {
		ID    string      `json:"id"`
		Value interface{} `json:"value,omitempty"`
	}{ID: resourceID, Value: details})
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	builder := tx.AuditLog.Create().SetAction(action).SetResource(resource).SetDetails(string(detailBytes)).SetCreatedAt(now)
	if userID != "" {
		builder.SetUserId(userID)
	}
	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

func runTransaction(ctx context.Context, client *ent.Client, operation func(*ent.Tx) error) error {
	if client == nil {
		return errors.New("membership: ent client is nil")
	}
	if operation == nil {
		return errors.New("membership: transaction operation is nil")
	}
	var lastErr error
	for attempt := 0; attempt < 20; attempt++ {
		tx, err := client.Tx(ctx)
		if err == nil {
			err = operation(tx)
			if err == nil {
				err = tx.Commit()
			} else {
				err = rollbackWithCause(tx, err)
			}
		}
		if err == nil {
			return nil
		}
		lastErr = err
		if !isBusyError(err) || attempt == 19 {
			return err
		}
		if err := waitForRetry(ctx, attempt+1); err != nil {
			return err
		}
	}
	return fmt.Errorf("transaction retry limit: %w", lastErr)
}

func rollbackWithCause(tx *ent.Tx, cause error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
		return errors.Join(cause, fmt.Errorf("rollback transaction: %w", rollbackErr))
	}
	return cause
}

func isBusyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database table is locked")
}

func waitForRetry(ctx context.Context, attempt int) error {
	timer := time.NewTimer(time.Duration(attempt) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait for transaction retry: %w", ctx.Err())
	}
}

func pageCount(total, limit int) int {
	if total == 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
