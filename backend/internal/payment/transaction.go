package payment

import (
	"context"
	databaseSQL "database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

func runTransaction(ctx context.Context, client *ent.Client, operation func(*ent.Tx) error) error {
	if client == nil || operation == nil {
		return errors.New("payment: invalid transaction dependencies")
	}
	var lastErr error
	for attempt := 0; attempt < 20; attempt++ {
		tx, err := client.Tx(ctx)
		if err == nil {
			err = operation(tx)
			if err == nil {
				err = tx.Commit()
			} else if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, databaseSQL.ErrTxDone) {
				err = errors.Join(err, rollbackErr)
			}
		}
		if err == nil {
			return nil
		}
		lastErr = err
		if !strings.Contains(strings.ToLower(err.Error()), "locked") || attempt == 19 {
			return err
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("payment transaction retry: %w", ctx.Err())
		}
	}
	return fmt.Errorf("payment transaction retry limit: %w", lastErr)
}
