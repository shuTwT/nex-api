package storage

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

func TestStorageRejectsPathLikeOriginalFilename(t *testing.T) {
	storage := newTestStorage(t)
	if _, err := storage.Save(context.Background(), bytes.NewReader(minimalPNG()), "../asset.png", "image/png"); !errors.Is(err, ErrInvalidFilename) {
		t.Fatalf("error = %v", err)
	}
	outside := filepath.Join(t.TempDir(), "outside.png")
	if err := os.WriteFile(outside, minimalPNG(), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(storage.directory, "link.png")); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	err := storage.ServeFile(rec, httptest.NewRequest(http.MethodGet, "/api/upload/link.png", nil), "link.png")
	if !errors.Is(err, ErrSymlink) {
		t.Fatalf("symlink error = %v", err)
	}
}
func TestUploadGeneratesDistinctNamesForDuplicateOriginalNames(t *testing.T) {
	storage := newTestStorage(t)
	first, err := storage.Save(context.Background(), bytes.NewReader(minimalPNG()), "same.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := storage.Save(context.Background(), bytes.NewReader(minimalPNG()), "same.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if first.Filename == second.Filename {
		t.Fatalf("duplicate names: %q", first.Filename)
	}
}
func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	storage, err := NewStorage(config.Upload{Directory: t.TempDir(), MaxBytes: MaxUploadBytes, CreateOnStart: true})
	if err != nil {
		t.Fatal(err)
	}
	return storage
}
func minimalPNG() []byte {
	return append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, 60)...)
}
