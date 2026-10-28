package httpkit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Credits int `json:"credits"`
	}

	var decoded payload
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"credits":1000}`))
	if err := DecodeJSON(request, &decoded); err != nil || decoded.Credits != 1000 {
		t.Fatalf("DecodeJSON() = %+v, %v", decoded, err)
	}

	for name, body := range map[string]string{
		"wrong type":    `{"credits":"1000"}`,
		"unknown field": `{"credits":1000,"extra":true}`,
		"multiple":      `{"credits":1000} {"credits":2000}`,
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			if err := DecodeJSON(request, &payload{}); !errors.Is(err, ErrValidation) {
				t.Fatalf("DecodeJSON() error = %v", err)
			}
		})
	}
}

func TestWritePaginated(t *testing.T) {
	recorder := httptest.NewRecorder()
	WritePaginated(recorder, []string{"item"}, 2, 10, 21)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	for _, expected := range []string{`"code":0`, `"total":21`, `"totalPages":3`, `"list":["item"]`} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("body = %s", recorder.Body.String())
		}
	}
}
