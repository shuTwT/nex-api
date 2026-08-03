package httpkit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/infra/logger"
)

func TestWriteData_writes_contract_success_response(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := WriteData(rec, http.StatusOK, map[string]any{"name": "nex", "enabled": true}); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Success bool            `json:"success"`
		Data    map[string]any  `json:"data"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || !payload.Success || payload.Data["name"] != "nex" || payload.Data["enabled"] != true {
		t.Fatalf("unexpected payload %#v", payload)
	}
	if len(payload.Error) != 0 && string(payload.Error) != "null" {
		t.Fatalf("expected omitted error, got %s", payload.Error)
	}
}
func TestWriteEnvelope_writes_contract_pagination_response(t *testing.T) {
	rec := httptest.NewRecorder()
	envelope := NewSuccessEnvelope([]string{"first", "second"})
	envelope.Pagination = &Pagination{Page: 2, PageSize: 25, Total: 51, TotalPages: 3}
	if err := WriteEnvelope(rec, http.StatusPartialContent, envelope); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Success    bool                       `json:"success"`
		Data       []string                   `json:"data"`
		Pagination map[string]json.RawMessage `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusPartialContent || !payload.Success || len(payload.Data) != 2 {
		t.Fatalf("unexpected payload %#v", payload)
	}
	for field, want := range map[string]int{"page": 2, "limit": 25, "total": 51, "totalPages": 3} {
		var got int
		if err := json.Unmarshal(payload.Pagination[field], &got); err != nil || got != want {
			t.Fatalf("pagination %s = %d, %v", field, got, err)
		}
	}
	if _, ok := payload.Pagination["page_size"]; ok {
		t.Fatal("pagination must not contain page_size")
	}
}
func TestWriteError_writes_message_without_internal_cause(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(logger.WithRequestID(req.Context(), "test-request-id"))
	apiError := NewAPIError(http.StatusBadRequest, "invalid_request", "request is invalid", errors.New("database secret"))
	if err := WriteError(rec, req, apiError); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Success bool            `json:"success"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	var message string
	if err := json.Unmarshal(payload.Error, &message); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest || payload.Success || message != apiError.Message || strings.Contains(message, "database secret") {
		t.Fatalf("unexpected error response: %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(logger.RequestIDHeader) != "test-request-id" {
		t.Fatalf("request ID = %q", rec.Header().Get(logger.RequestIDHeader))
	}
}
func TestWriteError_preserves_validation_status_and_contract_shape(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	if err := WriteError(rec, req, NewValidationError(FieldError{Field: "name", Reason: "required"})); err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Success bool            `json:"success"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	var message string
	if err := json.Unmarshal(payload.Error, &message); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest || payload.Success || message != "request validation failed" {
		t.Fatalf("unexpected validation response: %d %s", rec.Code, rec.Body.String())
	}
}
