package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/internal/infra/config"
	appstorage "github.com/shuTwT/nex-api/internal/infra/storage"
	serviceauth "github.com/shuTwT/nex-api/internal/service/auth"
	serviceupload "github.com/shuTwT/nex-api/internal/service/upload"
)

type uploadEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		Type     string `json:"type"`
	} `json:"data"`
}

func TestUploadAcceptsPNGAndServesBytesWithCacheHeaders(t *testing.T) {
	handler := NewHandler(newTestStorage(t))
	png := minimalPNG()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, multipartRequest(t, "asset.png", "image/png", png))
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d: %s", rec.Code, rec.Body.String())
	}
	var response uploadEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Data.URL == "" || response.Data.Filename == "" || response.Data.Size != int64(len(png)) || response.Data.Type != "image/png" {
		t.Fatalf("response = %+v", response)
	}
	served := httptest.NewRecorder()
	handler.ServeHTTP(served, httptest.NewRequest(http.MethodGet, response.Data.URL, nil))
	if served.Code != http.StatusOK || !bytes.Equal(served.Body.Bytes(), png) || served.Header().Get("Content-Type") != "image/png" || served.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" || served.Header().Get("Content-Length") != "68" {
		t.Fatalf("serve = %d %v", served.Code, served.Header())
	}
}
func TestUploadAcceptsExactly10MiB(t *testing.T) {
	content := append(minimalPNG(), bytes.Repeat([]byte{0}, int(appstorage.MaxUploadBytes)-len(minimalPNG()))...)
	rec := httptest.NewRecorder()
	NewHandler(newTestStorage(t)).ServeHTTP(rec, multipartRequest(t, "boundary.png", "image/png", content))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
}
func TestUploadRoutesRequireUserForPostAndKeepFileGetPublic(t *testing.T) {
	storage := newTestStorage(t)
	mux := chi.NewRouter()
	if err := RegisterRoutes(mux, storage); err != nil {
		t.Fatal(err)
	}
	unauthorized := httptest.NewRecorder()
	mux.ServeHTTP(unauthorized, multipartRequest(t, "asset.png", "image/png", minimalPNG()))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized = %d", unauthorized.Code)
	}
	pathPost := httptest.NewRecorder()
	request := multipartRequest(t, "asset.png", "image/png", minimalPNG())
	request.URL.Path = "/api/upload/extra.png"
	mux.ServeHTTP(pathPost, request)
	if pathPost.Code != http.StatusMethodNotAllowed {
		t.Fatalf("path POST = %d", pathPost.Code)
	}
	request = multipartRequest(t, "asset.png", "image/png", minimalPNG()).WithContext(serviceauth.WithAuthContext(context.Background(), serviceauth.AuthContext{User: serviceauth.User{ID: "user-1", Role: "user"}}))
	uploaded := httptest.NewRecorder()
	mux.ServeHTTP(uploaded, request)
	if uploaded.Code != http.StatusOK {
		t.Fatalf("upload = %d", uploaded.Code)
	}
	var response uploadEnvelope
	if err := json.Unmarshal(uploaded.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	served := httptest.NewRecorder()
	mux.ServeHTTP(served, httptest.NewRequest(http.MethodGet, response.Data.URL, nil))
	if served.Code != http.StatusOK || !bytes.Equal(served.Body.Bytes(), minimalPNG()) {
		t.Fatalf("public serve = %d", served.Code)
	}
}
func TestUploadRejectsOversizedMultipartWithoutLeavingTemporaryFiles(t *testing.T) {
	service := newTestStorage(t)
	tooLarge := append(minimalPNG(), bytes.Repeat([]byte{0}, int(appstorage.MaxUploadBytes))...)
	rec := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(rec, multipartRequest(t, "large.png", "image/png", tooLarge))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
func TestUploadRejectsUnsupportedMIMEAndPathLikeFilename(t *testing.T) {
	service := newTestStorage(t)
	rec := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(rec, multipartRequest(t, "notes.txt", "text/plain", []byte("not an image")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
func TestServeRejectsTraversalSymlinkAndReturns404ForMissing(t *testing.T) {
	handler := NewHandler(newTestStorage(t))
	for _, path := range []string{"/api/upload/../outside.png", "/api/upload/%2e%2e%2foutside.png", "/api/upload/missing.png"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound || bytes.Equal(rec.Body.Bytes(), minimalPNG()) {
			t.Fatalf("path %q status=%d", path, rec.Code)
		}
	}
}
func newTestStorage(t *testing.T) *serviceupload.Service {
	t.Helper()
	value, err := appstorage.NewStorage(config.Upload{Directory: t.TempDir(), MaxBytes: appstorage.MaxUploadBytes, CreateOnStart: true})
	if err != nil {
		t.Fatal(err)
	}
	return serviceupload.NewService(value)
}
func multipartRequest(t *testing.T, filename, mimeType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
func minimalPNG() []byte {
	return append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, 60)...)
}
