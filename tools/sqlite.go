package tools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"modernc.org/sqlite"
)

func copySQLiteDatabase(ctx context.Context, source, destination string) (err error) {
	if err := requireRegularSource(source); err != nil {
		return err
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect destination: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("create destination parent: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".sqlite-backup-*")
	if err != nil {
		return fmt.Errorf("create temporary database: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary database: %w", err)
	}
	if err := os.Remove(temporaryPath); err != nil {
		return fmt.Errorf("prepare temporary database: %w", err)
	}
	defer func() {
		if cleanupErr := removeIfPresent(temporaryPath); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	db, err := sql.Open("sqlite", sqliteReadOnlyDSN(source))
	if err != nil {
		return fmt.Errorf("open source database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	defer func() {
		err = errors.Join(err, db.Close())
	}()
	connection, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire source connection: %w", err)
	}
	defer func() {
		err = errors.Join(err, connection.Close())
	}()
	lockDB, err := sql.Open("sqlite", sqliteWriteLockDSN(source))
	if err != nil {
		return fmt.Errorf("open write lock database: %w", err)
	}
	lockDB.SetMaxOpenConns(1)
	lockDB.SetMaxIdleConns(1)
	defer func() {
		err = errors.Join(err, lockDB.Close())
	}()
	lockConnection, err := lockDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire write lock connection: %w", err)
	}
	defer func() {
		err = errors.Join(err, lockConnection.Close())
	}()
	writeLockHeld := false
	defer func() {
		if writeLockHeld {
			_, rollbackErr := lockConnection.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
			err = errors.Join(err, rollbackErr)
		}
	}()
	if _, err := lockConnection.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("stop sqlite writes: %w", err)
	}
	writeLockHeld = true

	if err := connection.Raw(func(driverConnection any) error {
		backuper, ok := driverConnection.(interface {
			NewBackup(string) (*sqlite.Backup, error)
		})
		if !ok {
			return errors.New("sqlite driver does not expose online backup")
		}
		backup, err := backuper.NewBackup(temporaryPath)
		if err != nil {
			return fmt.Errorf("start online backup: %w", err)
		}
		more, stepErr := backup.Step(-1)
		if stepErr != nil {
			return fmt.Errorf("copy sqlite pages: %w", stepErr)
		}
		if more {
			return errors.New("sqlite online backup did not finish")
		}
		destinationConnection, commitErr := backup.Commit()
		if destinationConnection != nil {
			commitErr = errors.Join(commitErr, destinationConnection.Close())
		}
		if commitErr != nil {
			return fmt.Errorf("finish online backup: %w", commitErr)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := syncFile(temporaryPath); err != nil {
		return fmt.Errorf("sync copied database: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("publish copied database: %w", err)
	}
	if _, err := lockConnection.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("resume sqlite writes: %w", err)
	}
	writeLockHeld = false
	return syncDirectory(filepath.Dir(destination))
}

func openSQLiteReadOnly(path string) (*sql.DB, error) {
	if err := requireRegularSource(path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", sqliteReadOnlyDSN(path))
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

func sqliteReadOnlyDSN(path string) string {
	return sqliteDSN(path, true)
}

func sqliteWriteLockDSN(path string) string {
	return sqliteDSN(path, false)
}
