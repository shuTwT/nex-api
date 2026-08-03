package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation failed")
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Cause      error
}

func NewAPIError(statusCode int, code string, message string, cause error) *APIError {
	return &APIError{StatusCode: statusCode, Code: code, Message: message, Cause: cause}
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

type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ValidationError struct {
	Fields []FieldError
}

func NewValidationError(fields ...FieldError) *ValidationError {
	return &ValidationError{Fields: append([]FieldError(nil), fields...)}
}

func (e *ValidationError) Error() string { return ErrValidation.Error() }

func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

type ErrorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id,omitempty"`
	Fields    []FieldError `json:"fields,omitempty"`
}

func (e ErrorBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Message)
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

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

func FromError(err error) *APIError {
	if err == nil {
		return nil
	}
	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError
	}
	var validationError *ValidationError
	switch {
	case errors.As(err, &validationError):
		return NewAPIError(http.StatusBadRequest, "validation_error", "request validation failed", err)
	case errors.Is(err, ErrNotFound):
		return NewAPIError(http.StatusNotFound, "not_found", "resource not found", err)
	case errors.Is(err, ErrUnauthorized):
		return NewAPIError(http.StatusUnauthorized, "unauthorized", "authentication required", err)
	case errors.Is(err, ErrForbidden):
		return NewAPIError(http.StatusForbidden, "forbidden", "access denied", err)
	case errors.Is(err, ErrConflict):
		return NewAPIError(http.StatusConflict, "conflict", "resource conflict", err)
	case errors.Is(err, context.Canceled):
		return NewAPIError(499, "request_canceled", "request canceled", err)
	case errors.Is(err, context.DeadlineExceeded):
		return NewAPIError(http.StatusGatewayTimeout, "request_timeout", "request timed out", err)
	default:
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return NewAPIError(http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", err)
		}
		return NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", err)
	}
}

func WriteData[T any](w http.ResponseWriter, status int, data T) error {
	return WriteEnvelope(w, status, NewSuccessEnvelope(data))
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	apiError := FromError(err)
	if apiError == nil {
		apiError = NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
	envelope := NewErrorEnvelope(apiError)
	if requestID := RequestID(r.Context()); requestID != "" {
		w.Header().Set(RequestIDHeader, requestID)
	}
	envelope.Error.RequestID = RequestID(r.Context())
	if validationError := asValidationError(err); validationError != nil {
		envelope.Error.Fields = append([]FieldError(nil), validationError.Fields...)
	}
	return WriteEnvelope(w, apiError.StatusCode, envelope)
}

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
