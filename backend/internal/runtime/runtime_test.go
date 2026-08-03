package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/config"
)

func TestNewServer_injects_handler_and_health_routes(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}),
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}

	// When
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected healthy response, got %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/injected", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected injected handler status 202, got %d", rec.Code)
	}
}

func TestServer_recovery_returns_typed_error_envelope_and_request_id(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("test panic")
		}),
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	rec := httptest.NewRecorder()

	// When
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if rec.Header().Get(RequestIDHeader) != "test-request-id" {
		t.Fatalf("expected request ID header, got %q", rec.Header().Get(RequestIDHeader))
	}
	if !strings.Contains(rec.Body.String(), `"error":"internal server error"`) {
		t.Fatalf("expected internal error envelope, got %s", rec.Body.String())
	}
}

func TestWriteData_writes_contract_success_response(t *testing.T) {
	// Given
	rec := httptest.NewRecorder()
	data := map[string]any{"name": "nex", "enabled": true}

	// When
	if err := WriteData(rec, http.StatusOK, data); err != nil {
		t.Fatalf("write success response: %v", err)
	}

	// Then
	var payload struct {
		Success bool            `json:"success"`
		Data    map[string]any  `json:"data"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode success response: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !payload.Success {
		t.Fatal("expected success to be true")
	}
	if payload.Data["name"] != "nex" || payload.Data["enabled"] != true {
		t.Fatalf("expected data to be preserved, got %#v", payload.Data)
	}
	if len(payload.Error) != 0 && string(payload.Error) != "null" {
		t.Fatalf("expected error to be omitted or null, got %s", payload.Error)
	}
}

func TestWriteEnvelope_writes_contract_pagination_response(t *testing.T) {
	// Given
	rec := httptest.NewRecorder()
	envelope := NewSuccessEnvelope([]string{"first", "second"})
	envelope.Pagination = &Pagination{Page: 2, PageSize: 25, Total: 51, TotalPages: 3}

	// When
	if err := WriteEnvelope(rec, http.StatusPartialContent, envelope); err != nil {
		t.Fatalf("write paginated response: %v", err)
	}

	// Then
	var payload struct {
		Success    bool                       `json:"success"`
		Data       []string                   `json:"data"`
		Pagination map[string]json.RawMessage `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode paginated response: %v", err)
	}
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("expected status 206, got %d", rec.Code)
	}
	if !payload.Success {
		t.Fatal("expected success to be true")
	}
	if len(payload.Data) != 2 || payload.Data[0] != "first" || payload.Data[1] != "second" {
		t.Fatalf("expected data to be preserved, got %#v", payload.Data)
	}
	expectedPagination := map[string]int{"page": 2, "limit": 25, "total": 51, "totalPages": 3}
	if len(payload.Pagination) != len(expectedPagination) {
		t.Fatalf("expected only contract pagination fields, got %#v", payload.Pagination)
	}
	for field, expected := range expectedPagination {
		var got int
		if err := json.Unmarshal(payload.Pagination[field], &got); err != nil {
			t.Fatalf("decode pagination field %q: %v", field, err)
		}
		if got != expected {
			t.Errorf("pagination %s: got %d, want %d", field, got, expected)
		}
	}
	if _, ok := payload.Pagination["page_size"]; ok {
		t.Error("pagination must not contain page_size")
	}
	if _, ok := payload.Pagination["total_pages"]; ok {
		t.Error("pagination must not contain total_pages")
	}
}

func TestWriteError_writes_message_without_internal_cause(t *testing.T) {
	// Given
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(WithRequestID(req.Context(), "test-request-id"))
	apiError := NewAPIError(http.StatusBadRequest, "invalid_request", "request is invalid", errors.New("database secret"))

	// When
	if err := WriteError(rec, req, apiError); err != nil {
		t.Fatalf("write error response: %v", err)
	}

	// Then
	var payload struct {
		Success bool            `json:"success"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if payload.Success {
		t.Fatal("expected success to be false")
	}
	var message string
	if err := json.Unmarshal(payload.Error, &message); err != nil {
		t.Fatalf("expected error to be a string: %v", err)
	}
	if message != apiError.Message {
		t.Fatalf("expected APIError.Message %q, got %q", apiError.Message, message)
	}
	if strings.Contains(message, "database secret") {
		t.Fatal("error response leaked internal cause")
	}
	if rec.Header().Get(RequestIDHeader) != "test-request-id" {
		t.Fatalf("expected request ID header, got %q", rec.Header().Get(RequestIDHeader))
	}
}

func TestWriteError_preserves_validation_status_and_contract_shape(t *testing.T) {
	// Given
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	validationErr := NewValidationError(FieldError{Field: "name", Reason: "required"})

	// When
	if err := WriteError(rec, req, validationErr); err != nil {
		t.Fatalf("write validation response: %v", err)
	}

	// Then
	var payload struct {
		Success bool            `json:"success"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if payload.Success {
		t.Fatal("expected success to be false")
	}
	var message string
	if err := json.Unmarshal(payload.Error, &message); err != nil {
		t.Fatalf("expected validation error to be a string: %v", err)
	}
	if message != "request validation failed" {
		t.Fatalf("expected validation message, got %q", message)
	}
}

func TestServer_readiness_reports_failed_dependency(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Readiness: []DependencyCheck{{
			Name:  "database",
			Check: func(context.Context) error { return errors.New("database unavailable") },
		}},
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}

	// When
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected readiness status 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"database":"failed"`) {
		t.Fatalf("expected failed database check, got %s", rec.Body.String())
	}
}
