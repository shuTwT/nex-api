package tools

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func sqliteDSN(path string, queryOnly bool) string {
	uri := url.URL{Scheme: "file", Path: path}
	query := url.Values{}
	query.Set("_busy_timeout", "5000")
	query.Set("_foreign_keys", "true")
	if queryOnly {
		query.Set("_query_only", "true")
	}
	uri.RawQuery = query.Encode()
	return uri.String()
}

func ResolveDatabasePath(value string) (string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if raw == "" {
		return filepath.Clean("prisma/dev.db"), nil
	}
	if !strings.HasPrefix(raw, "file:") {
		if strings.Contains(raw, "://") {
			return "", errors.New("DATABASE_URL must be a SQLite file URL")
		}
		return filepath.Clean(raw), nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse SQLite URL: %w", err)
	}
	path := parsed.Path
	if path == "" {
		path = parsed.Opaque
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		return "", errors.New("SQLite URL host must be empty or localhost")
	}
	if path == "" {
		return "", errors.New("SQLite URL has no database path")
	}
	return filepath.Clean(path), nil
}

func absolutePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

func requireRegularSource(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("symlinks are not accepted")
	}
	if !info.Mode().IsRegular() {
		return errors.New("path is not a regular file")
	}
	return nil
}

func createNewDirectory(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("path already exists: %s", path)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.MkdirAll(path, 0o700)
}

func syncFile(path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return file.Sync()
}

func syncDirectory(path string) (err error) {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()
	return directory.Sync()
}

func removeIfPresent(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func pathWithin(parent, target string) bool {
	relative, err := filepath.Rel(parent, target)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)))
}
