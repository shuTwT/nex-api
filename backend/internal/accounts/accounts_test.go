package accounts

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

func TestTokenView_neverMarshalsStoredToken(t *testing.T) {
	// Given
	entity := &ent.ApiToken{ID: "token-1", Name: "build", Token: "sk_secret-value", Permissions: "read", IsActive: true}

	// When
	body, err := json.Marshal(tokenView(entity))

	// Then
	if err != nil {
		t.Fatalf("marshal token view: %v", err)
	}
	if strings.Contains(string(body), entity.Token) || strings.Contains(string(body), `"token"`) {
		t.Fatalf("token view exposed secret: %s", body)
	}
}

func TestPageInfo_normalizesAndCalculatesPages(t *testing.T) {
	// Given
	request := PageRequest{Page: 0, Size: 0}

	// When
	got := pageInfo(request, 21)

	// Then
	if got.Page != 1 || got.PageSize != 10 || got.TotalPages != 3 {
		t.Fatalf("page info = %+v", got)
	}
}

func TestRequestMetadata_isValidJSONAndUsesRequestContext(t *testing.T) {
	// Given
	request := httptest.NewRequest("PUT", "/api/tokens/id", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("User-Agent", "accounts-test")

	// When
	metadata := requestMetadata(request)

	// Then
	if metadata.IP != "192.0.2.10" || metadata.UserAgent != "accounts-test" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if !json.Valid([]byte(metadata.Metadata)) {
		t.Fatalf("metadata is not JSON: %s", metadata.Metadata)
	}
}

func TestTokenView_preservesOptionalTimesWithoutRawToken(t *testing.T) {
	// Given
	expires := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	entity := &ent.ApiToken{ID: "token-1", ExpiresAt: expires, LastUsedAt: expires}

	// When
	view := tokenView(entity)

	// Then
	if view.ExpiresAt == nil || !view.ExpiresAt.Equal(expires) || view.LastUsedAt == nil || !view.LastUsedAt.Equal(expires) {
		t.Fatalf("optional token dates = %+v", view)
	}
}

func TestWriteAuditCSV_quotesCommasAndNewlines(t *testing.T) {
	// Given
	entity := &ent.AuditLog{CreatedAt: time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC), Action: "create, user", Resource: "users", Details: "line 1\nline 2", Level: "info", Status: "success"}
	var output bytes.Buffer

	// When
	if err := writeAuditCSV(&output, []*ent.AuditLog{entity}); err != nil {
		t.Fatalf("write CSV: %v", err)
	}

	// Then
	if !strings.Contains(output.String(), `"create, user"`) || !strings.Contains(output.String(), "\"line 1\nline 2\"") {
		t.Fatalf("CSV did not quote fields: %q", output.String())
	}
}
