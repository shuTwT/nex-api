package stats

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/api"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/mcpservice"
)

const defaultBatchSize = 100

type SyncService struct {
	Store     *Store
	Database  *ent.Client
	BatchSize int
}

func NewSyncService(store *Store, database *ent.Client) (*SyncService, error) {
	if store == nil {
		return nil, errors.New("stats: store is nil")
	}
	if database == nil {
		return nil, errors.New("stats: database is nil")
	}
	return &SyncService{Store: store, Database: database, BatchSize: defaultBatchSize}, nil
}

func (s *SyncService) Sync(ctx context.Context) error {
	if ctx == nil {
		return errors.New("stats: context is nil")
	}
	snapshot, err := s.Store.Snapshot(ctx)
	if err != nil {
		return fmt.Errorf("snapshot stats: %w", err)
	}
	if err := SyncToDatabase(ctx, s.Database, snapshot, s.BatchSize); err != nil {
		return fmt.Errorf("sync stats to database: %w", err)
	}
	return nil
}

func SyncToDatabase(ctx context.Context, database *ent.Client, snapshot Snapshot, batchSize int) error {
	if ctx == nil {
		return errors.New("stats: context is nil")
	}
	if database == nil {
		return errors.New("stats: database is nil")
	}
	if batchSize < 1 {
		batchSize = defaultBatchSize
	}
	apiAliases := sortedKeys(snapshot.APIs)
	for start := 0; start < len(apiAliases); start += batchSize {
		end := min(start+batchSize, len(apiAliases))
		if err := syncAPIBatch(ctx, database, snapshot.APIs, apiAliases[start:end]); err != nil {
			return fmt.Errorf("sync API batch %d-%d: %w", start, end, err)
		}
	}
	mcpIdentifiers := sortedKeys(snapshot.MCPs)
	for start := 0; start < len(mcpIdentifiers); start += batchSize {
		end := min(start+batchSize, len(mcpIdentifiers))
		if err := syncMCPBatch(ctx, database, snapshot.MCPs, mcpIdentifiers[start:end]); err != nil {
			return fmt.Errorf("sync MCP batch %d-%d: %w", start, end, err)
		}
	}
	return nil
}

func syncAPIBatch(ctx context.Context, database *ent.Client, counts map[string]int64, aliases []string) error {
	tx, err := database.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	for _, alias := range aliases {
		value, err := intCount(counts[alias])
		if err != nil {
			return rollback(tx, fmt.Errorf("API %q count: %w", alias, err))
		}
		entity, err := tx.Api.Query().Where(api.Alias(alias)).Only(ctx)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			return rollback(tx, fmt.Errorf("query API %q: %w", alias, err))
		}
		if _, err := tx.Api.UpdateOneID(entity.ID).SetCallCount(value).Save(ctx); err != nil {
			return rollback(tx, fmt.Errorf("update API %q: %w", alias, err))
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func syncMCPBatch(ctx context.Context, database *ent.Client, counts map[string]int64, identifiers []string) error {
	tx, err := database.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	for _, identifier := range identifiers {
		value, err := intCount(counts[identifier])
		if err != nil {
			return rollback(tx, fmt.Errorf("MCP %q count: %w", identifier, err))
		}
		entity, err := tx.McpService.Query().Where(mcpservice.Identifier(identifier)).Only(ctx)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			return rollback(tx, fmt.Errorf("query MCP %q: %w", identifier, err))
		}
		if _, err := tx.McpService.UpdateOneID(entity.ID).SetCallCount(value).Save(ctx); err != nil {
			return rollback(tx, fmt.Errorf("update MCP %q: %w", identifier, err))
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func sortedKeys(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func intCount(value int64) (int, error) {
	maxInt := int64(^uint(0) >> 1)
	if value < 0 || value > maxInt {
		return 0, fmt.Errorf("count %d does not fit in int", value)
	}
	return int(value), nil
}

func rollback(tx *ent.Tx, operationErr error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errors.Join(operationErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
	}
	return operationErr
}
