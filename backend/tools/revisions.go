package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ariga.io/atlas/sql/migrate"
)

type revisionStore struct {
	db *sql.DB
}

func (s *revisionStore) Ident() *migrate.TableIdent {
	return &migrate.TableIdent{Name: "atlas_schema_revisions"}
}

func (s *revisionStore) ensure(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "atlas_schema_revisions" (
			version TEXT NOT NULL PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			type INTEGER NOT NULL DEFAULT 2,
			applied INTEGER NOT NULL DEFAULT 0,
			total INTEGER NOT NULL DEFAULT 0,
			executed_at DATETIME NOT NULL,
			execution_time INTEGER NOT NULL DEFAULT 0,
			error TEXT NOT NULL DEFAULT '',
			error_stmt TEXT NOT NULL DEFAULT '',
			hash TEXT NOT NULL DEFAULT '',
			partial_hashes TEXT NOT NULL DEFAULT '',
			operator_version TEXT NOT NULL DEFAULT ''
		)`)
	return err
}

func (s *revisionStore) ReadRevisions(ctx context.Context) ([]*migrate.Revision, error) {
	rows, err := s.db.QueryContext(ctx, revisionSelect+" ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	revisions := make([]*migrate.Revision, 0)
	for rows.Next() {
		revision, err := scanRevision(rows)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, revision)
	}
	return revisions, rows.Err()
}

func (s *revisionStore) ReadRevision(ctx context.Context, version string) (*migrate.Revision, error) {
	revision, err := scanRevision(s.db.QueryRowContext(ctx, revisionSelect+" WHERE version = ?", version))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, migrate.ErrRevisionNotExist
	}
	return revision, err
}

func (s *revisionStore) WriteRevision(ctx context.Context, revision *migrate.Revision) error {
	if revision == nil {
		return errors.New("revision is nil")
	}
	partialHashes, err := json.Marshal(revision.PartialHashes)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO "atlas_schema_revisions" (
			version, description, type, applied, total, executed_at, execution_time,
			error, error_stmt, hash, partial_hashes, operator_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(version) DO UPDATE SET
			description = excluded.description,
			type = excluded.type,
			applied = excluded.applied,
			total = excluded.total,
			executed_at = excluded.executed_at,
			execution_time = excluded.execution_time,
			error = excluded.error,
			error_stmt = excluded.error_stmt,
			hash = excluded.hash,
			partial_hashes = excluded.partial_hashes,
			operator_version = excluded.operator_version`,
		revision.Version, revision.Description, int64(revision.Type), revision.Applied, revision.Total,
		revision.ExecutedAt.UTC().Format(time.RFC3339Nano), revision.ExecutionTime.Nanoseconds(),
		revision.Error, revision.ErrorStmt, revision.Hash, string(partialHashes), revision.OperatorVersion)
	return err
}

func (s *revisionStore) DeleteRevision(ctx context.Context, version string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM \"atlas_schema_revisions\" WHERE version = ?", version)
	return err
}

const revisionSelect = `
	SELECT version, description, CAST(type AS INTEGER), applied, total,
		CAST(executed_at AS TEXT), execution_time, error, error_stmt, hash,
		partial_hashes, operator_version
	FROM "atlas_schema_revisions"`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRevision(scanner rowScanner) (*migrate.Revision, error) {
	var (
		revision                    migrate.Revision
		revisionType, executionTime int64
		executedAt, partialHashes   string
	)
	if err := scanner.Scan(
		&revision.Version, &revision.Description, &revisionType, &revision.Applied, &revision.Total,
		&executedAt, &executionTime, &revision.Error, &revision.ErrorStmt, &revision.Hash,
		&partialHashes, &revision.OperatorVersion,
	); err != nil {
		return nil, err
	}
	revision.Type = migrate.RevisionType(revisionType)
	revision.ExecutionTime = time.Duration(executionTime)
	parsedTime, err := parseRevisionTime(executedAt)
	if err != nil {
		return nil, err
	}
	revision.ExecutedAt = parsedTime
	if partialHashes != "" && partialHashes != "null" {
		if err := json.Unmarshal([]byte(partialHashes), &revision.PartialHashes); err != nil {
			return nil, err
		}
	}
	return &revision, nil
}

func parseRevisionTime(value string) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05.999999999"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse revision time %q", value)
}

var _ migrate.RevisionReadWriter = (*revisionStore)(nil)
