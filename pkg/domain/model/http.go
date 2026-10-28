package model

import (
	"encoding/json"
	"fmt"

	"github.com/shuTwT/nex-api/pkg/constant"
)

// APIError is the HTTP projection of a service business error.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Cause      error
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

// ErrorBody is serialized as a plain string (its message) to keep the wire
// contract stable; request_id is only echoed through the header.
type ErrorBody struct {
	Code      string                `json:"code"`
	Message   string                `json:"message"`
	RequestID string                `json:"request_id,omitempty"`
	Fields    []constant.FieldError `json:"fields,omitempty"`
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
