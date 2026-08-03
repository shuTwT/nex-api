package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	serviceupload "github.com/shuTwT/nex-api/backend/internal/service/upload"
)

const multipartOverhead int64 = 1 << 20

type Handler struct {
	storage *serviceupload.Service
}

func NewHandler(storage *serviceupload.Service) *Handler { return &Handler{storage: storage} }

func RegisterRoutes(mux chi.Router, storage *serviceupload.Service) error {
	if mux == nil {
		return errors.New("upload: route mux is nil")
	}
	if storage == nil {
		return errors.New("upload: storage is nil")
	}
	handler := NewHandler(storage)
	mux.Method(http.MethodPost, "/api/upload", middleware.RequireUser(http.HandlerFunc(handler.UploadRoutePost)))
	mux.Method(http.MethodGet, "/api/upload/{filename}", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		handler.UploadFilenameRouteGet(writer, request, filenameFromRequest(request))
	}))
	return nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if h == nil || h.storage == nil {
		writeError(writer, request, errors.New("upload: storage is unavailable"))
		return
	}
	switch request.Method {
	case http.MethodPost:
		h.UploadRoutePost(writer, request)
	case http.MethodGet:
		h.UploadFilenameRouteGet(writer, request, filenameFromRequest(request))
	default:
		writer.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		writeError(writer, request, appRuntime.NewAPIError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil))
	}
}

func (h *Handler) UploadRoutePost(writer http.ResponseWriter, request *http.Request) {
	if h == nil || h.storage == nil {
		writeError(writer, request, errors.New("upload: storage is unavailable"))
		return
	}
	if request.ContentLength > h.storage.MaxBytes()+multipartOverhead {
		writeError(writer, request, serviceupload.ErrFileTooLarge)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, h.storage.MaxBytes()+multipartOverhead)
	if err := request.ParseMultipartForm(h.storage.MaxBytes()); err != nil {
		removeMultipartForm(request)
		writeError(writer, request, err)
		return
	}
	file, header, err := request.FormFile("file")
	if err != nil {
		removeErr := removeMultipartForm(request)
		if removeErr != nil {
			err = errors.Join(err, removeErr)
		}
		writeError(writer, request, err)
		return
	}
	metadata, saveErr := h.storage.Save(request.Context(), file, header.Filename, header.Header.Get("Content-Type"))
	closeErr := file.Close()
	removeErr := removeMultipartForm(request)
	if saveErr != nil {
		writeError(writer, request, saveErr)
		return
	}
	if closeErr != nil || removeErr != nil {
		writeError(writer, request, errors.Join(closeErr, removeErr))
		return
	}
	if err := appRuntime.WriteData(writer, http.StatusOK, metadata); err != nil {
		return
	}
}

func (h *Handler) UploadFilenameRouteGet(writer http.ResponseWriter, request *http.Request, filename string) {
	if h == nil || h.storage == nil {
		writeError(writer, request, errors.New("upload: storage is unavailable"))
		return
	}
	h.serve(writer, request, filename)
}

func (h *Handler) serve(writer http.ResponseWriter, request *http.Request, filename string) {
	content, err := h.storage.OpenFile(filename)
	if err != nil {
		writeError(writer, request, err)
		return
	}
	defer content.Body.Close()
	writer.Header().Set("Content-Type", content.ContentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", content.Size))
	writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = io.Copy(writer, content.Body)
}

func filenameFromRequest(request *http.Request) string {
	if filename := chi.URLParam(request, "filename"); filename != "" {
		return filename
	}
	const prefix = "/api/upload/"
	if path, ok := strings.CutPrefix(request.URL.Path, prefix); ok {
		return path
	}
	return ""
}

func removeMultipartForm(request *http.Request) error {
	if request.MultipartForm == nil {
		return nil
	}
	return request.MultipartForm.RemoveAll()
}

func writeError(writer http.ResponseWriter, request *http.Request, err error) {
	apiError := classifyError(err)
	if writeErr := appRuntime.WriteError(writer, request, apiError); writeErr != nil {
		return
	}
}

func classifyError(err error) error {
	if err == nil {
		return appRuntime.NewAPIError(http.StatusInternalServerError, "upload_failed", "upload failed", nil)
	}
	var maxBytesError *http.MaxBytesError
	switch {
	case errors.Is(err, serviceupload.ErrFileTooLarge), errors.As(err, &maxBytesError):
		return appRuntime.NewAPIError(http.StatusBadRequest, "file_too_large", "file size exceeds the 10MB limit", err)
	case errors.Is(err, serviceupload.ErrInvalidFilename), errors.Is(err, serviceupload.ErrInvalidMIME), errors.Is(err, serviceupload.ErrUnsupportedMIME), errors.Is(err, multipart.ErrMessageTooLarge):
		return appRuntime.NewAPIError(http.StatusBadRequest, "invalid_file", "file is invalid", err)
	case errors.Is(err, serviceupload.ErrFileNotFound), errors.Is(err, serviceupload.ErrSymlink), errors.Is(err, appRuntime.ErrNotFound), errors.Is(err, os.ErrNotExist):
		return appRuntime.NewAPIError(http.StatusNotFound, "not_found", "file not found", err)
	case isMethodNotAllowed(err):
		return err
	default:
		return appRuntime.NewAPIError(http.StatusInternalServerError, "upload_failed", "upload failed", err)
	}
}

func isMethodNotAllowed(err error) bool {
	var apiError *appRuntime.APIError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusMethodNotAllowed
}
