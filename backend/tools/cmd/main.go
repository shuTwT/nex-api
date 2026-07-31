package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shuTwT/nex-api/backend/tools"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: go run ./tools/cmd <backup|migrate|verify|restore|pipeline> [flags]")
	}
	switch args[0] {
	case "backup":
		return runBackup(ctx, args[1:], output, errorsOutput)
	case "migrate":
		return runMigrate(ctx, args[1:], output, errorsOutput)
	case "verify":
		return runVerify(ctx, args[1:], output, errorsOutput)
	case "restore":
		return runRestore(ctx, args[1:], output, errorsOutput)
	case "pipeline":
		return runPipeline(ctx, args[1:], output, errorsOutput)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runBackup(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	flags := flag.NewFlagSet("backup", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	database := flags.String("database", "", "SQLite file path or DATABASE_URL")
	uploads := flags.String("uploads", "data/upload", "upload directory")
	destination := flags.String("destination", "backup", "new backup directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	databasePath, err := tools.ResolveDatabasePath(*database)
	if err != nil {
		return err
	}
	manifest, err := tools.CreateBackup(ctx, tools.BackupConfig{
		SourceDatabase: databasePath,
		SourceUploads:  *uploads,
		Destination:    *destination,
	})
	if err != nil {
		return err
	}
	return writeJSON(output, manifest)
}

func runMigrate(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	database := flags.String("database", "", "isolated SQLite file path or DATABASE_URL")
	migrations := flags.String("migrations", defaultMigrationsDirectory(), "Atlas migration directory")
	baseline := flags.String("baseline", "20260730160000", "Atlas baseline version for the legacy schema")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*database) == "" {
		return errors.New("migrate requires --database pointing to an isolated backup database")
	}
	databasePath, err := tools.ResolveDatabasePath(*database)
	if err != nil {
		return err
	}
	config := tools.MigrationConfig{DatabasePath: databasePath, MigrationsDir: *migrations, Baseline: *baseline}
	if err := tools.ApplyMigrations(ctx, config); err != nil {
		return err
	}
	if err := tools.VerifyMigrations(ctx, config); err != nil {
		return err
	}
	return writeJSON(output, map[string]string{"database": databasePath, "status": "verified"})
}

func runVerify(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	database := flags.String("database", "", "SQLite file path or DATABASE_URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	databasePath, err := tools.ResolveDatabasePath(*database)
	if err != nil {
		return err
	}
	report, err := tools.VerifyDatabase(ctx, databasePath)
	if err != nil {
		return err
	}
	return writeJSON(output, report)
}

func runRestore(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	flags := flag.NewFlagSet("restore", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	backup := flags.String("backup", "backup", "backup directory")
	destination := flags.String("destination", "rollback", "new rollback directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if err := tools.RestoreBackup(ctx, tools.RestoreConfig{BackupDirectory: *backup, Destination: *destination}); err != nil {
		return err
	}
	return writeJSON(output, map[string]string{"destination": *destination, "status": "restored"})
}

func runPipeline(ctx context.Context, args []string, output, errorsOutput io.Writer) error {
	flags := flag.NewFlagSet("pipeline", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	database := flags.String("database", "", "SQLite file path or DATABASE_URL")
	uploads := flags.String("uploads", "data/upload", "upload directory")
	backup := flags.String("backup", "backup", "new backup directory")
	rollback := flags.String("rollback", "rollback", "new rollback directory")
	migrations := flags.String("migrations", defaultMigrationsDirectory(), "Atlas migration directory")
	baseline := flags.String("baseline", "20260730160000", "Atlas baseline version for the legacy schema")
	if err := flags.Parse(args); err != nil {
		return err
	}
	databasePath, err := tools.ResolveDatabasePath(*database)
	if err != nil {
		return err
	}
	manifest, err := tools.CreateBackup(ctx, tools.BackupConfig{
		SourceDatabase: databasePath,
		SourceUploads:  *uploads,
		Destination:    *backup,
	})
	if err != nil {
		return err
	}
	migrationConfig := tools.MigrationConfig{DatabasePath: manifest.DatabasePath, MigrationsDir: *migrations, Baseline: *baseline}
	if err := tools.ApplyMigrations(ctx, migrationConfig); err != nil {
		return err
	}
	if err := tools.VerifyMigrations(ctx, migrationConfig); err != nil {
		return err
	}
	integrity, err := tools.VerifyDatabase(ctx, manifest.DatabasePath)
	if err != nil {
		return err
	}
	if err := tools.RestoreBackup(ctx, tools.RestoreConfig{BackupDirectory: *backup, Destination: *rollback}); err != nil {
		return err
	}
	rollbackIntegrity, err := tools.VerifyDatabase(ctx, filepath.Join(filepath.Clean(*rollback), "database.sqlite"))
	if err != nil {
		return fmt.Errorf("verify rollback: %w", err)
	}
	return writeJSON(output, struct {
		Backup            tools.BackupManifest  `json:"backup"`
		MigratedIntegrity tools.IntegrityReport `json:"migratedIntegrity"`
		RollbackIntegrity tools.IntegrityReport `json:"rollbackIntegrity"`
		RollbackDirectory string                `json:"rollbackDirectory"`
	}{
		Backup:            manifest,
		MigratedIntegrity: integrity,
		RollbackIntegrity: rollbackIntegrity,
		RollbackDirectory: filepath.Clean(*rollback),
	})
}

func writeJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func defaultMigrationsDirectory() string {
	if _, err := os.Stat(filepath.Join("backend", "migrations")); err == nil {
		return filepath.Join("backend", "migrations")
	}
	return "migrations"
}
