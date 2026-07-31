package stats

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/config"
	"github.com/shuTwT/nex-api/backend/internal/cron"
)

func TestKeyMatrix_usesVersionedNamespaces_andPreservesLegacyAPIKeys(t *testing.T) {
	matrix := NewKeyMatrix()
	hour := time.Date(2026, 7, 30, 14, 37, 0, 0, time.FixedZone("test", 8*60*60))

	if got, want := matrix.APIRequests("weather"), "v1:stats:api:weather:requests"; got != want {
		t.Fatalf("API key = %q, want %q", got, want)
	}
	if got, want := matrix.UserAPICredits("u-1", "weather", hour), "v1:usage:user:u-1:api:weather:credits:1785391200"; got != want {
		t.Fatalf("user usage key = %q, want %q", got, want)
	}
	if got, want := legacyAPIKey("weather"), "api:request:count:weather"; got != want {
		t.Fatalf("legacy API key = %q, want %q", got, want)
	}
	if got, want := legacyUserAPIKey("u-1", "weather"), "user:api:request:count:u-1:weather"; got != want {
		t.Fatalf("legacy user API key = %q, want %q", got, want)
	}
}

func TestKeyMatrix_MCPKeys_useIdentifierWithoutLegacyAliasPrefix(t *testing.T) {
	matrix := NewKeyMatrix()
	if got, want := matrix.MCPRequests("weather-mcp"), "v1:stats:mcp:weather-mcp:requests"; got != want {
		t.Fatalf("MCP key = %q, want %q", got, want)
	}
	if got, want := matrix.UserMCPRequests("u-1", "weather-mcp"), "v1:stats:user:u-1:mcp:weather-mcp:requests"; got != want {
		t.Fatalf("user MCP key = %q, want %q", got, want)
	}
}

func TestStore_legacyMCPCanonicalKey_migratesGlobalAndUserAliases(t *testing.T) {
	store := &Store{matrix: NewKeyMatrix()}

	global, ok := store.legacyMCPCanonicalKey("api:request:count:mcp:weather")
	if !ok || global != "v1:stats:mcp:weather:requests" {
		t.Fatalf("global migration = (%q, %t)", global, ok)
	}
	user, ok := store.legacyMCPCanonicalKey("user:api:request:count:u-1:mcp:weather")
	if !ok || user != "v1:stats:user:u-1:mcp:weather:requests" {
		t.Fatalf("user migration = (%q, %t)", user, ok)
	}
}

func TestSyncToDatabase_intCount_rejects_negative_counts(t *testing.T) {
	if _, err := intCount(-1); err == nil {
		t.Fatal("negative count was accepted")
	}
	if value, err := intCount(42); err != nil || value != 42 {
		t.Fatalf("positive count = (%d, %v)", value, err)
	}
}

func TestSyncStatsHandler_requiresBearerSecret_whenEnabled(t *testing.T) {
	called := false
	handler, err := cron.NewSyncStatsHandler(config.Cron{Enabled: true, Secret: "cron-secret"}, func(context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	unauthorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/cron/sync-stats", nil)
	handler.ServeHTTP(unauthorized, request)
	if unauthorized.Code != http.StatusUnauthorized || called {
		t.Fatalf("unauthorized response = %d, sync called = %t", unauthorized.Code, called)
	}

	authorized := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/cron/sync-stats", nil)
	request.Header.Set("Authorization", "Bearer cron-secret")
	handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK || !called {
		t.Fatalf("authorized response = %d, sync called = %t", authorized.Code, called)
	}
}

func TestSyncStatsHandler_rejectsEnabledConfig_withoutSecret(t *testing.T) {
	_, err := cron.NewSyncStatsHandler(config.Cron{Enabled: true}, func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("enabled cron accepted an empty secret")
	}
}
