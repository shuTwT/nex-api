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
	"github.com/shuTwT/nex-api/pkg/constant"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

// Business error types, re-exported for callers that used to import them from
// the old runtime package. The HTTP wire types are aliases of pkg/domain/model.
type (
	FieldError      = serviceerror.FieldError
	ValidationError = serviceerror.ValidationError
	APIError        = model.APIError
	ErrorData       = model.ErrorData
	PageData[T any] = model.PageData[T]
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

// SuccessMsg is the msg of every successful response.
const SuccessMsg = "success"

func NewSuccessEnvelope[T any](data T) Envelope[T] {
	return Envelope[T]{Code: constant.CodeOK, Msg: SuccessMsg, Data: &data}
}

func NewErrorEnvelope(apiError *APIError) Envelope[ErrorData] {
	if apiError == nil {
		return Envelope[ErrorData]{Code: constant.CodeInternal, Msg: "internal server error"}
	}
	return Envelope[ErrorData]{Code: constant.CodeFor(apiError.Code, apiError.StatusCode), Msg: apiError.Message}
}

// WriteData writes a success envelope with the given status. A failed write
// (typically a disconnected client) has nobody to report to, so none is returned.
func WriteData[T any](w http.ResponseWriter, status int, data T) {
	_ = WriteEnvelope(w, status, NewSuccessEnvelope(data))
}

// WritePaginated writes a successful HTTP 200 envelope whose data is
// {list, total, page, limit, totalPages}.
func WritePaginated[T any](w http.ResponseWriter, list T, page, limit, total int) {
	pages := 0
	if limit > 0 {
		pages = total / limit
		if total%limit != 0 {
			pages++
		}
	}
	WriteData(w, http.StatusOK, PageData[T]{List: list, Total: total, Page: page, Limit: limit, TotalPages: pages})
}

// WriteFail writes an error response from a status, a slug and a message, for
// handlers that have no error value to map. Slugs map to numeric codes through
// constant.CodeFor.
func WriteFail(w http.ResponseWriter, r *http.Request, status int, slug, message string, cause error) {
	WriteError(w, r, NewAPIError(status, slug, message, cause))
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
	if requestID := logger.RequestID(r.Context()); requestID != "" {
		w.Header().Set(logger.RequestIDHeader, requestID)
	}
	envelope := NewErrorEnvelope(apiError)
	if validationError := asValidationError(err); validationError != nil {
		envelope.Data = &ErrorData{Fields: append([]FieldError(nil), validationError.Fields...)}
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
