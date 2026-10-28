package httpkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/internal/infra/config"
	"github.com/shuTwT/nex-api/internal/infra/logger"
	"github.com/shuTwT/nex-api/pkg/constant"
)

type testBody struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func decodeBody(t *testing.T, body []byte) testBody {
	t.Helper()
	var payload testBody
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return payload
}

func TestWriteData_writes_contract_success_response(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteData(rec, http.StatusOK, map[string]any{"name": "nex", "enabled": true})
	payload := decodeBody(t, rec.Body.Bytes())
	var data map[string]any
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || payload.Code != 0 || payload.Msg != SuccessMsg || data["name"] != "nex" || data["enabled"] != true {
		t.Fatalf("unexpected response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestWritePaginated_putsPaginationInsideData(t *testing.T) {
	rec := httptest.NewRecorder()
	WritePaginated(rec, []string{"first", "second"}, 2, 25, 51)
	payload := decodeBody(t, rec.Body.Bytes())
	var data struct {
		List       []string `json:"list"`
		Total      int      `json:"total"`
		Page       int      `json:"page"`
		Limit      int      `json:"limit"`
		TotalPages int      `json:"totalPages"`
	}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || payload.Code != 0 || len(data.List) != 2 || data.Total != 51 || data.Page != 2 || data.Limit != 25 || data.TotalPages != 3 {
		t.Fatalf("unexpected response: %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"pagination"`) {
		t.Fatalf("pagination must live inside data: %s", rec.Body.String())
	}
}

func TestWriteError_writesCodeAndMessageWithoutInternalCause(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(logger.WithRequestID(req.Context(), "test-request-id"))
	apiError := NewAPIError(http.StatusBadRequest, "invalid_request", "request is invalid", errors.New("database secret"))
	WriteError(rec, req, apiError)
	payload := decodeBody(t, rec.Body.Bytes())
	if rec.Code != http.StatusBadRequest || payload.Code != constant.CodeBadRequest || payload.Msg != apiError.Message || strings.Contains(rec.Body.String(), "database secret") {
		t.Fatalf("unexpected error response: %d %s", rec.Code, rec.Body.String())
	}
	if string(payload.Data) != "null" {
		t.Fatalf("data = %s, want null", payload.Data)
	}
	if rec.Header().Get(logger.RequestIDHeader) != "test-request-id" {
		t.Fatalf("request ID = %q", rec.Header().Get(logger.RequestIDHeader))
	}
}

func TestWriteError_validationCarriesFieldsInData(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	WriteError(rec, req, NewValidationError(FieldError{Field: "name", Reason: "required"}))
	payload := decodeBody(t, rec.Body.Bytes())
	var data struct {
		Fields []FieldError `json:"fields"`
	}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest || payload.Code != constant.CodeValidation || payload.Msg != "request validation failed" {
		t.Fatalf("unexpected validation response: %d %s", rec.Code, rec.Body.String())
	}
	if len(data.Fields) != 1 || data.Fields[0].Field != "name" || data.Fields[0].Reason != "required" {
		t.Fatalf("fields = %+v", data.Fields)
	}
}

func TestWriteError_codeAlwaysAgreesWithStatus(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusTeapot, http.StatusBadGateway} {
		rec := httptest.NewRecorder()
		WriteError(rec, httptest.NewRequest(http.MethodGet, "/", nil), NewAPIError(status, "slug_without_a_code", "msg", nil))
		payload := decodeBody(t, rec.Body.Bytes())
		if payload.Code/100 != status {
			t.Fatalf("status %d produced code %d", status, payload.Code)
		}
	}
}

func TestWriteFail_usesSlugCode(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteFail(rec, httptest.NewRequest(http.MethodPost, "/", nil), http.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误", nil)
	payload := decodeBody(t, rec.Body.Bytes())
	if rec.Code != http.StatusUnauthorized || payload.Code != constant.CodeInvalidCredentials || payload.Msg != "邮箱或密码错误" {
		t.Fatalf("unexpected response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestWriteError_logsInternalCauseWithRequestID(t *testing.T) {
	var buffer bytes.Buffer
	testLogger, err := logger.NewLoggerWithWriter(config.Log{Level: "info", Format: "text"}, &buffer)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/payment/methods", nil)
	req = req.WithContext(logger.WithLogger(logger.WithRequestID(req.Context(), "req-500"), testLogger))
	cause := errors.New("get payment settings: database is locked")
	rec := httptest.NewRecorder()
	WriteError(rec, req, cause)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), cause.Error()) {
		t.Fatalf("response leaked the cause: %s", rec.Body.String())
	}
	logged := buffer.String()
	if !strings.Contains(logged, "request failed") || !strings.Contains(logged, cause.Error()) || !strings.Contains(logged, "req-500") || !strings.Contains(logged, "/api/payment/methods") {
		t.Fatalf("expected cause chain in log, got %q", logged)
	}
}

func TestWriteError_doesNotLogBusinessErrors(t *testing.T) {
	var buffer bytes.Buffer
	testLogger, err := logger.NewLoggerWithWriter(config.Log{Level: "info", Format: "text"}, &buffer)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req = req.WithContext(logger.WithLogger(req.Context(), testLogger))
	rec := httptest.NewRecorder()
	WriteError(rec, req, NewAPIError(http.StatusNotFound, "not_found", "resource not found", errors.New("hidden detail")))
	if buffer.Len() != 0 {
		t.Fatalf("business errors must not be logged, got %q", buffer.String())
	}
}
