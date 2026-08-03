// Package apierror defines business errors produced by services. It deliberately
// contains no HTTP concepts; handler/httpkit maps Kind to an HTTP response.
package apierror

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation failed")
)

// Kind describes the business category of an Error. It is intentionally not a
// transport status; HTTP handlers decide how each kind is represented on the
// wire.
type Kind string

const (
	KindInvalidInput Kind = "invalid_input"
	KindNotFound     Kind = "not_found"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindConflict     Kind = "conflict"
	KindUnavailable  Kind = "unavailable"
)

// Error is a structured domain error with a stable application code.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Cause   error
}

func NewError(kind Kind, code, message string, cause error) *Error {
	return &Error{Kind: kind, Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return e.Code + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *Error) Unwrap() error {
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

func (e *ValidationError) Error() string        { return ErrValidation.Error() }
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

// KindOf classifies generic business errors without assigning HTTP statuses.
func KindOf(err error) Kind {
	var businessError *Error
	if errors.As(err, &businessError) {
		return businessError.Kind
	}
	var validationError *ValidationError
	switch {
	case errors.As(err, &validationError), errors.Is(err, ErrValidation):
		return KindInvalidInput
	case errors.Is(err, ErrNotFound):
		return KindNotFound
	case errors.Is(err, ErrUnauthorized):
		return KindUnauthorized
	case errors.Is(err, ErrForbidden):
		return KindForbidden
	case errors.Is(err, ErrConflict):
		return KindConflict
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return KindUnavailable
	default:
		return ""
	}
}
