package storage

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type FileContent struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

// OpenFile returns validated upload content without writing an HTTP response.
func (s *Storage) OpenFile(filename string) (content FileContent, err error) {
	if s == nil {
		return FileContent{}, errors.New("upload: storage is unavailable")
	}
	path, err := s.confinedPath(filename)
	if err != nil {
		if errors.Is(err, ErrInvalidFilename) {
			return FileContent{}, ErrFileNotFound
		}
		return FileContent{}, err
	}
	file, err := openWithoutFollowingSymlink(path)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return FileContent{}, ErrSymlink
		}
		if errors.Is(err, os.ErrNotExist) {
			return FileContent{}, ErrFileNotFound
		}
		return FileContent{}, fmt.Errorf("open upload: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return FileContent{}, fmt.Errorf("stat served upload: %w", err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return FileContent{}, ErrFileNotFound
	}
	prefix := make([]byte, 512)
	read, readErr := file.Read(prefix)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		_ = file.Close()
		return FileContent{}, fmt.Errorf("detect upload type: %w", readErr)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		return FileContent{}, fmt.Errorf("rewind upload: %w", err)
	}
	contentType := detectMIME(prefix[:read])
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = fallbackContentType(filename)
	}
	return FileContent{Body: file, ContentType: contentType, Size: info.Size()}, nil
}

func (s *Storage) ServeFile(writer http.ResponseWriter, request *http.Request, filename string) (err error) {
	if s == nil {
		return errors.New("upload: storage is unavailable")
	}
	path, err := s.confinedPath(filename)
	if err != nil {
		if errors.Is(err, ErrInvalidFilename) {
			return ErrFileNotFound
		}
		return err
	}
	file, err := openWithoutFollowingSymlink(path)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return ErrSymlink
		}
		if errors.Is(err, os.ErrNotExist) {
			return ErrFileNotFound
		}
		return fmt.Errorf("open upload: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close served upload: %w", closeErr)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat served upload: %w", err)
	}
	if !info.Mode().IsRegular() {
		return ErrFileNotFound
	}
	prefix := make([]byte, 512)
	read, readErr := file.Read(prefix)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fmt.Errorf("detect upload type: %w", readErr)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind upload: %w", err)
	}
	contentType := detectMIME(prefix[:read])
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = fallbackContentType(filename)
	}
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	writer.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
	http.ServeContent(writer, request, filename, info.ModTime(), file)
	return nil
}

func openWithoutFollowingSymlink(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		if closeErr := syscall.Close(fd); closeErr != nil {
			return nil, errors.Join(errors.New("create file handle"), closeErr)
		}
		return nil, errors.New("create file handle")
	}
	return file, nil
}

func fallbackContentType(filename string) string {
	extension := strings.ToLower(filepath.Ext(filename))
	if contentType := mime.TypeByExtension(extension); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func (s *Storage) confinedPath(filename string) (string, error) {
	if err := validateFilename(filename); err != nil {
		return "", err
	}
	if err := s.validateDirectory(); err != nil {
		return "", err
	}
	candidate := filepath.Join(s.directory, filename)
	relative, err := filepath.Rel(s.directory, candidate)
	if err != nil {
		return "", fmt.Errorf("confine upload path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("%s: %w", filename, ErrInvalidFilename)
	}
	if info, err := os.Lstat(candidate); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%s: %w", filename, ErrSymlink)
	}
	return candidate, nil
}

func (s *Storage) ensureDirectory() error {
	if err := rejectSymlinkComponents(s.directory); err != nil {
		return err
	}
	if err := os.MkdirAll(s.directory, 0o755); err != nil {
		return fmt.Errorf("make upload directory: %w", err)
	}
	return s.validateDirectory()
}

func (s *Storage) validateDirectory() error {
	info, err := os.Lstat(s.directory)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrSymlink
	}
	if !info.IsDir() {
		return fmt.Errorf("upload path is not a directory: %s", s.directory)
	}
	return rejectSymlinkComponents(s.directory)
}

func validateOriginalFilename(filename string) error {
	if filename == "" || len(filename) > 255 || filename == "." || filename == ".." {
		return ErrInvalidFilename
	}
	if strings.ContainsAny(filename, `/\\`) || strings.Contains(filename, "..") {
		return ErrInvalidFilename
	}
	for _, char := range filename {
		if char < 0x20 || char == 0x7f {
			return ErrInvalidFilename
		}
	}
	return nil
}

func validateFilename(filename string) error {
	if err := validateOriginalFilename(filename); err != nil {
		return err
	}
	if strings.Contains(filename, "%") {
		return ErrInvalidFilename
	}
	for _, char := range filename {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return ErrInvalidFilename
	}
	return nil
}

func rejectSymlinkComponents(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path components: %w", err)
	}
	current := filepath.VolumeName(absolute) + string(os.PathSeparator)
	remaining := strings.TrimPrefix(absolute, current)
	for component := range strings.SplitSeq(remaining, string(os.PathSeparator)) {
		if component == "" {
			continue
		}
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		if statErr != nil {
			return fmt.Errorf("inspect path component %s: %w", current, statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path component %s: %w", current, ErrSymlink)
		}
	}
	return nil
}

func canonicalDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	absolute = filepath.Clean(absolute)
	if info, statErr := os.Lstat(absolute); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", ErrSymlink
		}
		resolved, evalErr := filepath.EvalSymlinks(absolute)
		if evalErr != nil {
			return "", fmt.Errorf("canonicalize path: %w", evalErr)
		}
		return filepath.Clean(resolved), nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return "", fmt.Errorf("inspect path: %w", statErr)
	}

	missing := make([]string, 0, 2)
	existing := absolute
	for {
		_, statErr := os.Lstat(existing)
		if statErr == nil {
			break
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return "", fmt.Errorf("inspect path component: %w", statErr)
		}
		missing = append(missing, filepath.Base(existing))
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", os.ErrNotExist
		}
		existing = parent
	}
	resolved, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", fmt.Errorf("canonicalize parent: %w", err)
	}
	for index := len(missing) - 1; index >= 0; index-- {
		resolved = filepath.Join(resolved, missing[index])
	}
	return filepath.Clean(resolved), nil
}
