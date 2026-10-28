package model

import (
	"fmt"

	"github.com/shuTwT/nex-api/pkg/constant"
)

// APIError is the HTTP projection of a service business error. Code is the
// string slug used inside the code base; the numeric response code is derived
// from it (and the status) by constant.CodeFor when the response is written.
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

// ErrorData is the optional `data` payload of an error response. It is only
// present when there is structured detail, such as per-field validation errors.
type ErrorData struct {
	Fields []constant.FieldError `json:"fields,omitempty"`
}

// PageData is the `data` payload of a paginated response.
type PageData[T any] struct {
	List       T   `json:"list"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// Envelope is the uniform response wrapper: Code 0 means success, any other
// value is an entry of the table in pkg/constant (HTTP status * 100 + sequence).
type Envelope[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data *T     `json:"data"`
}
