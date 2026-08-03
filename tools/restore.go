package tools

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// RestoreConfig identifies a backup and a new rollback directory.
type RestoreConfig struct {
	BackupDirectory string
	Destination     string
}

// RestoreBackup copies a backup into a separate, newly-created rollback location.
func RestoreBackup(ctx context.Context, config RestoreConfig) (err error) {
	backupDirectory, err := absolutePath(config.BackupDirectory)
	if err != nil {
		return fmt.Errorf("resolve backup directory: %w", err)
	}
	destination, err := absolutePath(config.Destination)
	if err != nil {
		return fmt.Errorf("resolve restore destination: %w", err)
	}
	if pathWithin(backupDirectory, destination) || pathWithin(destination, backupDirectory) {
		return errors.New("restore destination overlaps the backup directory")
	}
	if err := requireDirectory(backupDirectory); err != nil {
		return fmt.Errorf("backup directory: %w", err)
	}
	if err := requireRegularSource(filepath.Join(backupDirectory, databaseFilename)); err != nil {
		return fmt.Errorf("backup database: %w", err)
	}
	if err := createNewDestinationParent(destination); err != nil {
		return fmt.Errorf("prepare restore destination: %w", err)
	}
	temporary, err := os.MkdirTemp(filepath.Dir(destination), ".restore-*")
	if err != nil {
		return fmt.Errorf("create restore staging directory: %w", err)
	}
	temporaryPath := temporary
	defer func() {
		if cleanupErr := os.RemoveAll(temporaryPath); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	if err := copySQLiteDatabase(ctx, filepath.Join(backupDirectory, databaseFilename), filepath.Join(temporaryPath, databaseFilename)); err != nil {
		return fmt.Errorf("restore sqlite database: %w", err)
	}
	if _, err := copyUploads(filepath.Join(backupDirectory, uploadsDirname), filepath.Join(temporaryPath, uploadsDirname)); err != nil {
		return fmt.Errorf("restore uploads: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("publish restore: %w", err)
	}
	return syncDirectory(filepath.Dir(destination))
}

func requireDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("symlinks are not accepted")
	}
	if !info.IsDir() {
		return errors.New("path is not a directory")
	}
	return nil
}

func createNewDestinationParent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("restore destination already exists: %s", path)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.MkdirAll(filepath.Dir(path), 0o700)
}
