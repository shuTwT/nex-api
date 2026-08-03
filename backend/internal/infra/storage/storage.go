package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

const MaxUploadBytes int64 = 10 << 20

var defaultAllowedTypes = map[string]struct{}{
	"image/gif":  {},
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

var mimeExtensions = map[string]string{
	"image/gif":  "gif",
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

var (
	ErrFileTooLarge    = errors.New("upload: file exceeds size limit")
	ErrInvalidFilename = errors.New("upload: invalid filename")
	ErrInvalidMIME     = errors.New("upload: invalid MIME type")
	ErrUnsupportedMIME = errors.New("upload: unsupported MIME type")
	ErrSymlink         = errors.New("upload: symlink is not allowed")
	ErrFileNotFound    = errors.New("upload: file not found")
)

type Metadata struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

type Storage struct {
	directory   string
	maxBytes    int64
	allowedType map[string]struct{}
}

func NewStorage(cfg config.Upload) (*Storage, error) {
	directory := strings.TrimSpace(cfg.Directory)
	if directory == "" {
		directory = filepath.Join("data", "upload")
	}
	absolute, err := canonicalDirectory(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve upload directory: %w", err)
	}

	maxBytes := cfg.MaxBytes
	if maxBytes <= 0 || maxBytes > MaxUploadBytes {
		maxBytes = MaxUploadBytes
	}
	allowedType, err := normalizeAllowedTypes(cfg.AllowedTypes)
	if err != nil {
		return nil, err
	}
	storage := &Storage{directory: absolute, maxBytes: maxBytes, allowedType: allowedType}
	if cfg.CreateOnStart {
		if err := storage.ensureDirectory(); err != nil {
			return nil, fmt.Errorf("create upload directory: %w", err)
		}
	}
	return storage, nil
}

func New(cfg config.Upload) (*Storage, error) { return NewStorage(cfg) }

func (s *Storage) Directory() string { return s.directory }

func (s *Storage) MaxBytes() int64 { return s.maxBytes }

func (s *Storage) Save(ctx context.Context, source io.Reader, originalFilename, declaredType string) (metadata Metadata, err error) {
	if ctx == nil {
		return Metadata{}, errors.New("upload: context is nil")
	}
	if source == nil {
		return Metadata{}, errors.New("upload: source is nil")
	}
	if err := validateOriginalFilename(originalFilename); err != nil {
		return Metadata{}, err
	}
	if err := s.ensureDirectory(); err != nil {
		return Metadata{}, fmt.Errorf("prepare upload directory: %w", err)
	}

	temporary, err := os.CreateTemp(s.directory, ".upload-*")
	if err != nil {
		return Metadata{}, fmt.Errorf("create temporary upload: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	closed := false
	defer func() {
		if !closed {
			if closeErr := temporary.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("close temporary upload: %w", closeErr)
			}
		}
		if !committed {
			if removeErr := os.Remove(temporaryPath); err == nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = fmt.Errorf("remove temporary upload: %w", removeErr)
			}
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return Metadata{}, fmt.Errorf("restrict temporary upload: %w", err)
	}

	size, prefix, err := copyUpload(ctx, temporary, source, s.maxBytes)
	if err != nil {
		return Metadata{}, err
	}
	detectedType := detectMIME(prefix)
	if err := s.validateMIME(declaredType, detectedType); err != nil {
		return Metadata{}, err
	}
	if err := temporary.Sync(); err != nil {
		return Metadata{}, fmt.Errorf("sync temporary upload: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return Metadata{}, fmt.Errorf("close temporary upload: %w", err)
	}
	closed = true

	filename := uuid.NewString() + "." + extensionForMIME(detectedType)
	target, err := s.confinedPath(filename)
	if err != nil {
		return Metadata{}, err
	}
	if _, err := os.Lstat(target); err == nil {
		return Metadata{}, fmt.Errorf("reserve generated upload filename: %w", os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Metadata{}, fmt.Errorf("inspect generated upload filename: %w", err)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return Metadata{}, fmt.Errorf("publish upload: %w", err)
	}
	committed = true
	return Metadata{URL: "/api/upload/" + filename, Filename: filename, Size: size, Type: detectedType}, nil
}

func (s *Storage) validateMIME(declaredType, detectedType string) error {
	declared, err := normalizeMIME(declaredType)
	if err != nil {
		return err
	}
	if _, ok := s.allowedType[detectedType]; !ok {
		return fmt.Errorf("%s: %w", detectedType, ErrUnsupportedMIME)
	}
	if declared != detectedType {
		return fmt.Errorf("declared %s, detected %s: %w", declared, detectedType, ErrInvalidMIME)
	}
	return nil
}

func normalizeAllowedTypes(values []string) (map[string]struct{}, error) {
	if len(values) == 0 {
		return cloneTypes(defaultAllowedTypes), nil
	}
	allowed := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized, err := normalizeMIME(value)
		if err != nil {
			return nil, fmt.Errorf("allowed MIME %q: %w", value, err)
		}
		allowed[normalized] = struct{}{}
	}
	return allowed, nil
}

func cloneTypes(values map[string]struct{}) map[string]struct{} {
	clone := make(map[string]struct{}, len(values))
	for value := range values {
		clone[value] = struct{}{}
	}
	return clone
}

func normalizeMIME(value string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil || mediaType == "" || strings.Contains(mediaType, "*") {
		return "", fmt.Errorf("%q: %w", value, ErrInvalidMIME)
	}
	return strings.ToLower(mediaType), nil
}

func detectMIME(prefix []byte) string {
	mediaType, _, err := mime.ParseMediaType(http.DetectContentType(prefix))
	if err != nil {
		return ""
	}
	return strings.ToLower(mediaType)
}

func extensionForMIME(mediaType string) string {
	if extension, ok := mimeExtensions[mediaType]; ok {
		return extension
	}
	return "bin"
}

func copyUpload(ctx context.Context, destination io.Writer, source io.Reader, maxBytes int64) (int64, []byte, error) {
	buffer := make([]byte, 32*1024)
	prefix := make([]byte, 0, 512)
	var size int64
	for {
		if err := ctx.Err(); err != nil {
			return 0, nil, fmt.Errorf("read upload: %w", err)
		}
		read, readErr := source.Read(buffer)
		if read > 0 {
			if size+int64(read) > maxBytes {
				return 0, nil, ErrFileTooLarge
			}
			written, writeErr := destination.Write(buffer[:read])
			if writeErr != nil {
				return 0, nil, fmt.Errorf("write upload: %w", writeErr)
			}
			if written != read {
				return 0, nil, io.ErrShortWrite
			}
			size += int64(read)
			if len(prefix) < 512 {
				remaining := min(512-len(prefix), read)
				prefix = append(prefix, buffer[:remaining]...)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return size, prefix, nil
			}
			return 0, nil, fmt.Errorf("read upload: %w", readErr)
		}
	}
}
