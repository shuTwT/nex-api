package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/shuTwT/nex-api/internal/handler/httpkit"
)

const maxJSONBodyBytes = 1 << 20

// DecodeJSON decodes exactly one JSON value, rejects unknown fields, and limits
// the request body to 1 MiB. Decode failures are returned as field validation errors.
func DecodeJSON(r *http.Request, destination any) error {
	if r == nil || r.Body == nil || destination == nil {
		return invalidJSONError()
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return invalidJSONError()
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return httpkit.NewValidationError(httpkit.FieldError{
			Field:  "body",
			Reason: "must contain exactly one JSON value",
		})
	}
	return nil
}

// DecodeJSONValue is the value-returning form of DecodeJSON.
func DecodeJSONValue[T any](r *http.Request) (T, error) {
	var value T
	err := DecodeJSON(r, &value)
	return value, err
}

func invalidJSONError() error {
	return httpkit.NewValidationError(httpkit.FieldError{
		Field:  "body",
		Reason: "invalid JSON",
	})
}
