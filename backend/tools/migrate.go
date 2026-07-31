package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"

	"ariga.io/atlas/sql/migrate"
	atlassqlite "ariga.io/atlas/sql/sqlite"
)

// MigrationConfig identifies an isolated database and its Atlas migration directory.
type MigrationConfig struct {
	DatabasePath  string
	MigrationsDir string
	Baseline      string
}

// MigrationReport describes applied and still-pending Atlas files.
type MigrationReport struct {
	Applied []string `json:"applied"`
	Pending []string `json:"pending"`
	Current string   `json:"current"`
}

// ApplyMigrations applies all pending Atlas migrations to the configured copy.
func ApplyMigrations(ctx context.Context, config MigrationConfig) (err error) {
	db, executor, revisions, err := newMigrationExecutor(ctx, config)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := pendingFiles(ctx, executor); err != nil {
		return fmt.Errorf("inspect pending migrations: %w", err)
	}
	if err := executor.ExecuteN(ctx, 0); err != nil {
		return fmt.Errorf("apply Atlas migrations: %w", err)
	}
	report, err := migrationStatus(ctx, executor, revisions)
	if err != nil {
		return err
	}
	if len(report.Pending) > 0 {
		return fmt.Errorf("Atlas left %d migration(s) pending", len(report.Pending))
	}
	return nil
}

// VerifyMigrations confirms the copied database is at the latest Atlas revision.
func VerifyMigrations(ctx context.Context, config MigrationConfig) (err error) {
	db, executor, revisions, err := newMigrationExecutor(ctx, config)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	report, err := migrationStatus(ctx, executor, revisions)
	if err != nil {
		return err
	}
	directory, err := migrate.NewLocalDir(config.MigrationsDir)
	if err != nil {
		return fmt.Errorf("open Atlas migration directory: %w", err)
	}
	files, err := directory.Files()
	if err != nil {
		return fmt.Errorf("read Atlas migration files: %w", err)
	}
	if len(files) == 0 {
		return errors.New("Atlas migration directory is empty")
	}
	if len(report.Pending) != 0 {
		return fmt.Errorf("Atlas migration verification found %d pending file(s)", len(report.Pending))
	}
	if report.Current != files[len(files)-1].Version() {
		return fmt.Errorf("Atlas is at %q, want %q", report.Current, files[len(files)-1].Version())
	}
	return nil
}

func newMigrationExecutor(ctx context.Context, config MigrationConfig) (*sql.DB, *migrate.Executor, *revisionStore, error) {
	if err := requireRegularSource(config.DatabasePath); err != nil {
		return nil, nil, nil, fmt.Errorf("migration database: %w", err)
	}
	info, err := os.Stat(config.MigrationsDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("migration directory: %w", err)
	}
	if !info.IsDir() {
		return nil, nil, nil, errors.New("migration path is not a directory")
	}
	db, err := sql.Open("sqlite", config.DatabasePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open migration database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("ping migration database: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	driver, err := atlassqlite.Open(db)
	if err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("create Atlas SQLite driver: %w", err)
	}
	directory, err := migrate.NewLocalDir(config.MigrationsDir)
	if err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("open Atlas migration directory: %w", err)
	}
	revisions := &revisionStore{db: db}
	if err := revisions.ensure(ctx); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("prepare Atlas revision table: %w", err)
	}
	options := make([]migrate.ExecutorOption, 0, 1)
	if config.Baseline != "" {
		options = append(options, migrate.WithBaselineVersion(config.Baseline))
	}
	executor, err := migrate.NewExecutor(driver, directory, revisions, options...)
	if err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("create Atlas executor: %w", err)
	}
	return db, executor, revisions, nil
}

func migrationStatus(ctx context.Context, executor *migrate.Executor, revisions *revisionStore) (MigrationReport, error) {
	pending, err := pendingFiles(ctx, executor)
	if err != nil {
		return MigrationReport{}, fmt.Errorf("read Atlas status: %w", err)
	}
	applied, err := revisions.ReadRevisions(ctx)
	if err != nil {
		return MigrationReport{}, fmt.Errorf("read Atlas revisions: %w", err)
	}
	pendingNames := make([]string, len(pending))
	for index := range pending {
		pendingNames[index] = pending[index].Name()
	}
	appliedNames := make([]string, len(applied))
	for index := range applied {
		appliedNames[index] = applied[index].Version
	}
	sort.Strings(appliedNames)
	current := ""
	if len(appliedNames) > 0 {
		current = appliedNames[len(appliedNames)-1]
	}
	return MigrationReport{Applied: appliedNames, Pending: pendingNames, Current: current}, nil
}

func pendingFiles(ctx context.Context, executor *migrate.Executor) ([]migrate.File, error) {
	pending, err := executor.Pending(ctx)
	if errors.Is(err, migrate.ErrNoPendingFiles) {
		return []migrate.File{}, nil
	}
	if err != nil {
		return nil, err
	}
	return pending, nil
}
