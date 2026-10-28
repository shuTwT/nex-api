// Package httpkit provides the unified JSON envelope, pagination and error
// response helpers for the handler layer, plus mapping for business errors
// defined by service/apierror. The wire types live in pkg/domain/model.
package httpkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/shuTwT/nex-api/internal/infra/logger"
	serviceerror "github.com/shuTwT/nex-api/internal/service/apierror"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

// Business error types, re-exported for callers that used to import them from
// the old runtime package. The HTTP wire types are aliases of pkg/domain/model.
type (
	FieldError      = serviceerror.FieldError
	ValidationError = serviceerror.ValidationError
	APIError        = model.APIError
	ErrorBody       = model.ErrorBody
	Pagination      = model.Pagination
	Envelope[T any] = model.Envelope[T]
)

var (
	ErrNotFound     = serviceerror.ErrNotFound
	ErrUnauthorized = serviceerror.ErrUnauthorized
	ErrForbidden    = serviceerror.ErrForbidden
	ErrConflict     = serviceerror.ErrConflict
	ErrValidation   = serviceerror.ErrValidation
)

func NewAPIError(statusCode int, code string, message string, cause error) *APIError {
	return &APIError{StatusCode: statusCode, Code: code, Message: message, Cause: cause}
}

func NewValidationError(fields ...FieldError) *ValidationError {
	return serviceerror.NewValidationError(fields...)
}

func FromError(err error) *APIError {
	if err == nil {
		return nil
	}
	var httpError *APIError
	if errors.As(err, &httpError) {
		return httpError
	}
	var businessError *serviceerror.Error
	if errors.As(err, &businessError) {
		return NewAPIError(statusForKind(businessError.Kind), businessError.Code, businessError.Message, err)
	}
	switch serviceerror.KindOf(err) {
	case serviceerror.KindInvalidInput:
		return NewAPIError(http.StatusBadRequest, "validation_error", "request validation failed", err)
	case serviceerror.KindNotFound:
		return NewAPIError(http.StatusNotFound, "not_found", "resource not found", err)
	case serviceerror.KindUnauthorized:
		return NewAPIError(http.StatusUnauthorized, "unauthorized", "authentication required", err)
	case serviceerror.KindForbidden:
		return NewAPIError(http.StatusForbidden, "forbidden", "access denied", err)
	case serviceerror.KindConflict:
		return NewAPIError(http.StatusConflict, "conflict", "resource conflict", err)
	case serviceerror.KindUnavailable:
		if errors.Is(err, context.Canceled) {
			return NewAPIError(499, "request_canceled", "request canceled", err)
		}
		return NewAPIError(http.StatusGatewayTimeout, "request_timeout", "request timed out", err)
	}
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return NewAPIError(http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", err)
	}
	return NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", err)
}

func statusForKind(kind serviceerror.Kind) int {
	switch kind {
	case serviceerror.KindInvalidInput:
		return http.StatusBadRequest
	case serviceerror.KindNotFound:
		return http.StatusNotFound
	case serviceerror.KindUnauthorized:
		return http.StatusUnauthorized
	case serviceerror.KindForbidden:
		return http.StatusForbidden
	case serviceerror.KindConflict:
		return http.StatusConflict
	case serviceerror.KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func NewSuccessEnvelope[T any](data T) Envelope[T] {
	return Envelope[T]{Success: true, Data: &data}
}

func NewErrorEnvelope(apiError *APIError) Envelope[struct{}] {
	if apiError == nil {
		return Envelope[struct{}]{Success: false, Error: &ErrorBody{}}
	}
	return Envelope[struct{}]{
		Success: false,
		Error:   &ErrorBody{Code: apiError.Code, Message: apiError.Message},
	}
}

// WriteData writes a success envelope with the given status. A failed write
// (typically a disconnected client) has nobody to report to, so none is returned.
func WriteData[T any](w http.ResponseWriter, status int, data T) {
	_ = WriteEnvelope(w, status, NewSuccessEnvelope(data))
}

// WritePaginated writes a successful HTTP 200 envelope with pagination metadata.
func WritePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := 0
	if limit > 0 {
		pages = total / limit
		if total%limit != 0 {
			pages++
		}
	}
	_ = WriteEnvelope(w, http.StatusOK, Envelope[T]{
		Success:    true,
		Data:       &data,
		Pagination: &Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages},
	})
}

// WriteError maps err to a compatible error envelope and writes it. Errors
// mapped to a 5xx response are logged with their full cause chain — without
// this the client only sees "internal server error" and the root cause is
// lost.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	apiError := FromError(err)
	if apiError == nil {
		apiError = NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
	if apiError.StatusCode >= http.StatusInternalServerError {
		logger.LoggerFrom(r.Context()).ErrorContext(r.Context(), "request failed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", apiError.StatusCode),
			slog.Any("error", err),
		)
	}
	envelope := NewErrorEnvelope(apiError)
	if requestID := logger.RequestID(r.Context()); requestID != "" {
		w.Header().Set(logger.RequestIDHeader, requestID)
	}
	envelope.Error.RequestID = logger.RequestID(r.Context())
	if validationError := asValidationError(err); validationError != nil {
		envelope.Error.Fields = append([]FieldError(nil), validationError.Fields...)
	}
	_ = WriteEnvelope(w, apiError.StatusCode, envelope)
}

// WriteEnvelope serializes the envelope as the JSON response body.
func WriteEnvelope[T any](w http.ResponseWriter, status int, envelope Envelope[T]) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		return fmt.Errorf("encode response envelope: %w", err)
	}
	return nil
}

func asValidationError(err error) *ValidationError {
	var validationError *ValidationError
	if errors.As(err, &validationError) {
		return validationError
	}
	return nil
}
