package upload

import (
	"context"
	"io"

	"github.com/shuTwT/nex-api/internal/infra/storage"
)

type Metadata struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

type FileContent struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

var (
	ErrFileTooLarge    = storage.ErrFileTooLarge
	ErrInvalidFilename = storage.ErrInvalidFilename
	ErrInvalidMIME     = storage.ErrInvalidMIME
	ErrUnsupportedMIME = storage.ErrUnsupportedMIME
	ErrSymlink         = storage.ErrSymlink
	ErrFileNotFound    = storage.ErrFileNotFound
)

type Service struct{ storage *storage.Storage }

func NewService(value *storage.Storage) *Service { return &Service{storage: value} }
func (s *Service) MaxBytes() int64 {
	if s == nil || s.storage == nil {
		return 0
	}
	return s.storage.MaxBytes()
}
func (s *Service) Save(ctx context.Context, source io.Reader, filename, contentType string) (Metadata, error) {
	value, err := s.storage.Save(ctx, source, filename, contentType)
	return Metadata{URL: value.URL, Filename: value.Filename, Size: value.Size, Type: value.Type}, err
}
func (s *Service) OpenFile(filename string) (FileContent, error) {
	value, err := s.storage.OpenFile(filename)
	return FileContent{Body: value.Body, ContentType: value.ContentType, Size: value.Size}, err
}
