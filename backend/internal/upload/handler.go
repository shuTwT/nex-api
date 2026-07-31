package upload

import (
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

const multipartOverhead int64 = 1 << 20

type Handler struct {
	storage *Storage
}

func NewHandler(storage *Storage) *Handler { return &Handler{storage: storage} }

func RegisterRoutes(mux *http.ServeMux, storage *Storage) error {
	if mux == nil {
		return errors.New("upload: route mux is nil")
	}
	if storage == nil {
		return errors.New("upload: storage is nil")
	}
	handler := NewHandler(storage)
	mux.Handle("/api/upload", http.HandlerFunc(handler.UploadRoutePost))
	mux.Handle("/api/upload/", handler)
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
		writeError(writer, request, runtime.NewAPIError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil))
	}
}

func (h *Handler) UploadRoutePost(writer http.ResponseWriter, request *http.Request) {
	if h == nil || h.storage == nil {
		writeError(writer, request, errors.New("upload: storage is unavailable"))
		return
	}
	if request.ContentLength > h.storage.maxBytes+multipartOverhead {
		writeError(writer, request, ErrFileTooLarge)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, h.storage.maxBytes+multipartOverhead)
	if err := request.ParseMultipartForm(h.storage.maxBytes); err != nil {
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
	if err := runtime.WriteData(writer, http.StatusOK, metadata); err != nil {
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
	if err := serveFile(writer, request, h.storage, filename); err != nil {
		writeError(writer, request, err)
	}
}

func filenameFromRequest(request *http.Request) string {
	if filename := request.PathValue("filename"); filename != "" {
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
	if writeErr := runtime.WriteError(writer, request, apiError); writeErr != nil {
		return
	}
}

func classifyError(err error) error {
	if err == nil {
		return runtime.NewAPIError(http.StatusInternalServerError, "upload_failed", "upload failed", nil)
	}
	var maxBytesError *http.MaxBytesError
	switch {
	case errors.Is(err, ErrFileTooLarge), errors.As(err, &maxBytesError):
		return runtime.NewAPIError(http.StatusBadRequest, "file_too_large", "file size exceeds the 10MB limit", err)
	case errors.Is(err, ErrInvalidFilename), errors.Is(err, ErrInvalidMIME), errors.Is(err, ErrUnsupportedMIME), errors.Is(err, multipart.ErrMessageTooLarge):
		return runtime.NewAPIError(http.StatusBadRequest, "invalid_file", "file is invalid", err)
	case errors.Is(err, ErrFileNotFound), errors.Is(err, ErrSymlink), errors.Is(err, runtime.ErrNotFound), errors.Is(err, os.ErrNotExist):
		return runtime.NewAPIError(http.StatusNotFound, "not_found", "file not found", err)
	case isMethodNotAllowed(err):
		return err
	default:
		return runtime.NewAPIError(http.StatusInternalServerError, "upload_failed", "upload failed", err)
	}
}

func isMethodNotAllowed(err error) bool {
	var apiError *runtime.APIError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusMethodNotAllowed
}
