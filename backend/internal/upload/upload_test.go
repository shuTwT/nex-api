package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/auth"
	"github.com/shuTwT/nex-api/backend/internal/config"
)

func TestUploadAcceptsPNGAndServesBytesWithCacheHeaders(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	handler := NewHandler(storage)
	png := minimalPNG()
	request := multipartRequest(t, "asset.png", "image/png", png)
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response responseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if !response.Success || response.Data.URL == "" || response.Data.Filename == "" {
		t.Fatalf("unexpected upload response: %+v", response)
	}
	if response.Data.Size != int64(len(png)) || response.Data.Type != "image/png" {
		t.Fatalf("unexpected upload metadata: %+v", response.Data)
	}

	serveRecorder := httptest.NewRecorder()
	serveRequest := httptest.NewRequest(http.MethodGet, response.Data.URL, nil)
	handler.ServeHTTP(serveRecorder, serveRequest)
	if serveRecorder.Code != http.StatusOK {
		t.Fatalf("serve status = %d, want %d", serveRecorder.Code, http.StatusOK)
	}
	if !bytes.Equal(serveRecorder.Body.Bytes(), png) {
		t.Fatal("served bytes differ from uploaded bytes")
	}
	if got := serveRecorder.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("content type = %q, want image/png", got)
	}
	if got := serveRecorder.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("cache control = %q", got)
	}
	if got := serveRecorder.Header().Get("Content-Length"); got != "68" {
		t.Fatalf("content length = %q", got)
	}
}

func TestUploadAcceptsExactly10MiB(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	content := append(minimalPNG(), bytes.Repeat([]byte{0}, int(MaxUploadBytes)-len(minimalPNG()))...)
	recorder := httptest.NewRecorder()

	// When
	NewHandler(storage).ServeHTTP(recorder, multipartRequest(t, "boundary.png", "image/png", content))

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("exact-limit status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func TestUploadRoutesRequireUserForPostAndKeepFileGetPublic(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, storage); err != nil {
		t.Fatal(err)
	}

	// When: an unauthenticated upload is rejected before parsing the body
	unauthenticated := httptest.NewRecorder()
	mux.ServeHTTP(unauthenticated, multipartRequest(t, "asset.png", "image/png", minimalPNG()))

	// Then
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated upload status = %d, want %d: %s", unauthenticated.Code, http.StatusUnauthorized, unauthenticated.Body.String())
	}

	// A POST to a file path must not fall through to the public file handler.
	pathUpload := httptest.NewRecorder()
	pathRequest := multipartRequest(t, "asset.png", "image/png", minimalPNG())
	pathRequest.URL.Path = "/api/upload/extra.png"
	mux.ServeHTTP(pathUpload, pathRequest)
	if pathUpload.Code != http.StatusMethodNotAllowed {
		t.Fatalf("file-path upload status = %d, want %d: %s", pathUpload.Code, http.StatusMethodNotAllowed, pathUpload.Body.String())
	}

	request := multipartRequest(t, "asset.png", "image/png", minimalPNG())
	request = request.WithContext(auth.WithAuthContext(request.Context(), auth.AuthContext{
		User: auth.User{ID: "user-1", Role: "user"},
	}))
	uploaded := httptest.NewRecorder()
	mux.ServeHTTP(uploaded, request)
	if uploaded.Code != http.StatusOK {
		t.Fatalf("authenticated upload status = %d, want %d: %s", uploaded.Code, http.StatusOK, uploaded.Body.String())
	}

	var response responseEnvelope
	if err := json.Unmarshal(uploaded.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}

	// The generated file URL remains public and does not need auth context.
	served := httptest.NewRecorder()
	mux.ServeHTTP(served, httptest.NewRequest(http.MethodGet, response.Data.URL, nil))
	if served.Code != http.StatusOK {
		t.Fatalf("public file status = %d, want %d: %s", served.Code, http.StatusOK, served.Body.String())
	}
	if !bytes.Equal(served.Body.Bytes(), minimalPNG()) {
		t.Fatal("public file bytes differ from uploaded bytes")
	}
}

func TestUploadRejectsOversizedMultipartWithoutLeavingTemporaryFiles(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	tooLarge := append(minimalPNG(), bytes.Repeat([]byte{0}, int(MaxUploadBytes))...)
	request := multipartRequest(t, "large.png", "image/png", tooLarge)
	recorder := httptest.NewRecorder()

	// When
	NewHandler(storage).ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("oversized status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	entries, err := os.ReadDir(storage.directory)
	if err != nil {
		t.Fatalf("read upload directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".upload-") {
			t.Fatalf("temporary upload remains: %s", entry.Name())
		}
	}
}

func TestUploadRejectsUnsupportedMIMEAndPathLikeFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		mimeType string
		body     []byte
	}{
		{name: "unsupported mime", filename: "notes.txt", mimeType: "text/plain", body: []byte("not an image")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			storage := newTestStorage(t)
			request := multipartRequest(t, test.filename, test.mimeType, test.body)
			recorder := httptest.NewRecorder()

			// When
			NewHandler(storage).ServeHTTP(recorder, request)

			// Then
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
			}
			entries, err := os.ReadDir(storage.directory)
			if err != nil {
				t.Fatalf("read upload directory: %v", err)
			}
			if len(entries) != 0 {
				t.Fatalf("rejected upload created files: %v", entries)
			}
		})
	}
}

func TestStorageRejectsPathLikeOriginalFilename(t *testing.T) {
	// Given
	storage := newTestStorage(t)

	// When
	_, err := storage.Save(context.Background(), bytes.NewReader(minimalPNG()), "../asset.png", "image/png")

	// Then
	if !errors.Is(err, ErrInvalidFilename) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidFilename)
	}
}

func TestServeRejectsTraversalSymlinkAndReturns404ForMissing(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	outside := filepath.Join(t.TempDir(), "outside.png")
	if err := os.WriteFile(outside, minimalPNG(), 0o600); err != nil {
		t.Fatalf("write outside fixture: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(storage.directory, "link.png")); err != nil {
		t.Fatalf("create symlink fixture: %v", err)
	}
	handler := NewHandler(storage)

	for _, path := range []string{"/api/upload/../outside.png", "/api/upload/%2e%2e%2foutside.png", "/api/upload/link.png", "/api/upload/missing.png"} {
		// When
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		// Then
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("path %q status = %d, want %d", path, recorder.Code, http.StatusNotFound)
		}
		if bytes.Equal(recorder.Body.Bytes(), minimalPNG()) {
			t.Fatalf("path %q served bytes outside the upload directory", path)
		}
	}
}

func TestUploadGeneratesDistinctNamesForDuplicateOriginalNames(t *testing.T) {
	// Given
	storage := newTestStorage(t)
	handler := NewHandler(storage)
	names := make([]string, 0, 2)
	for range 2 {
		request := multipartRequest(t, "same.png", "image/png", minimalPNG())
		recorder := httptest.NewRecorder()

		// When
		handler.ServeHTTP(recorder, request)

		// Then
		if recorder.Code != http.StatusOK {
			t.Fatalf("upload status = %d, want %d", recorder.Code, http.StatusOK)
		}
		var response responseEnvelope
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode upload response: %v", err)
		}
		names = append(names, response.Data.Filename)
	}
	if names[0] == names[1] {
		t.Fatalf("duplicate upload names: %q", names[0])
	}
}

type responseEnvelope struct {
	Success bool           `json:"success"`
	Data    uploadMetadata `json:"data"`
}

type uploadMetadata struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	storage, err := NewStorage(config.Upload{Directory: t.TempDir(), MaxBytes: MaxUploadBytes, CreateOnStart: true})
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return storage
}

func multipartRequest(t *testing.T, filename, mimeType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	partHeader.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("create multipart part: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("write multipart part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func minimalPNG() []byte {
	return append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, 60)...)
}
