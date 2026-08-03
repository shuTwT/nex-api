// Package httpkit provides the unified JSON envelope, pagination and error
// response helpers for the handler layer, plus mapping for business errors
// defined by service/apierror.
package httpkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/shuTwT/nex-api/backend/internal/middleware"
	serviceerror "github.com/shuTwT/nex-api/backend/internal/service/apierror"
)

// Business error types, re-exported for callers that used to import them from
// the old runtime package.
type (
	FieldError      = serviceerror.FieldError
	ValidationError = serviceerror.ValidationError
)

var (
	ErrNotFound     = serviceerror.ErrNotFound
	ErrUnauthorized = serviceerror.ErrUnauthorized
	ErrForbidden    = serviceerror.ErrForbidden
	ErrConflict     = serviceerror.ErrConflict
	ErrValidation   = serviceerror.ErrValidation
)

// APIError is the HTTP-layer projection of a service business error.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Cause      error
}

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

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return e.Code + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
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

// ErrorBody is serialized as a plain string (its message) to keep the wire
// contract stable; request_id is only echoed through the header.
type ErrorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id,omitempty"`
	Fields    []FieldError `json:"fields,omitempty"`
}

func (e ErrorBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Message)
}

// Pagination describes the paged result envelope fields.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Envelope is the uniform success/error response wrapper.
type Envelope[T any] struct {
	Success    bool        `json:"success"`
	Data       *T          `json:"data,omitempty"`
	Error      *ErrorBody  `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
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

// WriteData writes a success envelope with the given status.
func WriteData[T any](w http.ResponseWriter, status int, data T) error {
	return WriteEnvelope(w, status, NewSuccessEnvelope(data))
}

// WriteError maps err to a compatible error envelope and writes it.
func WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	apiError := FromError(err)
	if apiError == nil {
		apiError = NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
	envelope := NewErrorEnvelope(apiError)
	if requestID := middleware.RequestID(r.Context()); requestID != "" {
		w.Header().Set(middleware.RequestIDHeader, requestID)
	}
	envelope.Error.RequestID = middleware.RequestID(r.Context())
	if validationError := asValidationError(err); validationError != nil {
		envelope.Error.Fields = append([]FieldError(nil), validationError.Fields...)
	}
	return WriteEnvelope(w, apiError.StatusCode, envelope)
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
