package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_recovery_returns_typed_error_envelope_and_request_id(t *testing.T) {
	handler := RequestIDMiddleware(Recovery(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("test panic") })))
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set(RequestIDHeader, "test-request-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, request)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get(RequestIDHeader) != "test-request-id" {
		t.Fatalf("request ID = %q", rec.Header().Get(RequestIDHeader))
	}
	if !strings.Contains(rec.Body.String(), `"error":"internal server error"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
