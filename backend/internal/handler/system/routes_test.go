package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/ent/enttest"
	servicesystem "github.com/shuTwT/nex-api/backend/internal/service/system"
)

func TestHandler_Initialized_returnsAPIEnvelope(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/system.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	service, err := servicesystem.NewService(client)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/system/initialized", nil)
	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Initialized bool `json:"initialized"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Fatal("success = false, want true")
	}
	if body.Data.Initialized {
		t.Fatal("initialized = true, want false")
	}
	_ = context.Background()
}
