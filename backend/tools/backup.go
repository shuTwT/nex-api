package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	manifestFilename = "manifest.json"
	databaseFilename = "database.sqlite"
	uploadsDirname   = "upload"
)

// BackupConfig describes a source tree and a new backup directory.
type BackupConfig struct {
	SourceDatabase string
	SourceUploads  string
	Destination    string
}

// UploadChecksum records one copied regular upload file.
type UploadChecksum struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

// BackupManifest is written only after the copied database and uploads pass verification.
type BackupManifest struct {
	Version          int              `json:"version"`
	CreatedAt        string           `json:"createdAt"`
	SourceDatabase   string           `json:"sourceDatabase"`
	DatabasePath     string           `json:"databasePath"`
	UploadsDirectory string           `json:"uploadsDirectory"`
	DatabaseSHA256   string           `json:"databaseSha256"`
	Uploads          []UploadChecksum `json:"uploads"`
	IntegrityReport
}

// CreateBackup makes an isolated SQLite online backup and a non-symlink upload copy.
func CreateBackup(ctx context.Context, config BackupConfig) (BackupManifest, error) {
	sourceDatabase, err := absolutePath(config.SourceDatabase)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("resolve source database: %w", err)
	}
	sourceUploads, err := absolutePath(config.SourceUploads)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("resolve source uploads: %w", err)
	}
	destination, err := absolutePath(config.Destination)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("resolve backup destination: %w", err)
	}
	if pathWithin(sourceUploads, destination) || pathWithin(destination, sourceDatabase) {
		return BackupManifest{}, errors.New("backup destination overlaps a source path")
	}
	if err := requireRegularSource(sourceDatabase); err != nil {
		return BackupManifest{}, fmt.Errorf("source database: %w", err)
	}
	if err := createNewDirectory(destination); err != nil {
		return BackupManifest{}, fmt.Errorf("create backup directory: %w", err)
	}

	databasePath := filepath.Join(destination, databaseFilename)
	if err := copySQLiteDatabase(ctx, sourceDatabase, databasePath); err != nil {
		return BackupManifest{}, fmt.Errorf("copy sqlite database: %w", err)
	}
	uploadsDirectory := filepath.Join(destination, uploadsDirname)
	uploads, err := copyUploads(sourceUploads, uploadsDirectory)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("copy uploads: %w", err)
	}
	report, err := VerifyDatabase(ctx, databasePath)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("verify backup database: %w", err)
	}
	databaseSHA256, err := hashFile(databasePath)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("hash backup database: %w", err)
	}
	manifest := BackupManifest{
		Version:          1,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		SourceDatabase:   sourceDatabase,
		DatabasePath:     databasePath,
		UploadsDirectory: uploadsDirectory,
		DatabaseSHA256:   databaseSHA256,
		Uploads:          uploads,
		IntegrityReport:  report,
	}
	if err := writeManifest(filepath.Join(destination, manifestFilename), manifest); err != nil {
		return BackupManifest{}, fmt.Errorf("write backup manifest: %w", err)
	}
	return manifest, nil
}

func writeManifest(path string, manifest BackupManifest) (err error) {
	contents, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), ".manifest-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { err = errors.Join(err, removeIfPresent(temporaryPath)) }()
	if _, err := temporary.Write(contents); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
