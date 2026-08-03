package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceauth "github.com/shuTwT/nex-api/backend/internal/service/auth"
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
			if err := DecodeJSON(request, &payload{}); !errors.Is(err, httpkit.ErrValidation) {
				t.Fatalf("DecodeJSON() error = %v", err)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name   string
		user   serviceauth.User
		ok     bool
		status int
	}{
		{name: "admin", user: serviceauth.User{ID: "admin-1", Role: "admin"}, ok: true, status: http.StatusOK},
		{name: "user", user: serviceauth.User{ID: "user-1", Role: "user"}, status: http.StatusForbidden},
		{name: "anonymous", status: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.user.ID != "" {
				request = request.WithContext(serviceauth.WithAuthContext(request.Context(), serviceauth.AuthContext{User: test.user}))
			}
			recorder := httptest.NewRecorder()
			if ok := RequireAdmin(recorder, request); ok != test.ok {
				t.Fatalf("RequireAdmin() = %v", ok)
			}
			if recorder.Code != test.status {
				t.Fatalf("status = %d", recorder.Code)
			}
		})
	}
}

func TestPositiveInt(t *testing.T) {
	if value, err := PositiveInt("", 10); err != nil || value != 10 {
		t.Fatalf("fallback = %d, %v", value, err)
	}
	if value, err := PositiveInt("2", 10); err != nil || value != 2 {
		t.Fatalf("parsed = %d, %v", value, err)
	}
	if _, err := PositiveInt("0", 10); err == nil {
		t.Fatal("PositiveInt accepted zero")
	}
}

func TestWritePaginated(t *testing.T) {
	recorder := httptest.NewRecorder()
	WritePaginated(recorder, []string{"item"}, 2, 10, 21)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	for _, expected := range []string{`"success":true`, `"total":21`, `"totalPages":3`} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("body = %s", recorder.Body.String())
		}
	}
}
