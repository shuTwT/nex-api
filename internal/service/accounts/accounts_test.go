package accounts

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/shuTwT/nex-api/ent"
)

func TestTokenView_neverMarshalsStoredToken(t *testing.T) {
	entity := &ent.ApiToken{ID: "token-1", Name: "build", Token: "sk_secret-value", Permissions: "read", IsActive: true}
	body, err := json.Marshal(tokenView(entity))
	if err != nil {
		t.Fatalf("marshal token view: %v", err)
	}
	if strings.Contains(string(body), entity.Token) || strings.Contains(string(body), `"token"`) {
		t.Fatalf("token view exposed secret: %s", body)
	}
}

func TestPageInfo_normalizesAndCalculatesPages(t *testing.T) {
	got := pageInfo(PageRequest{Page: 0, Size: 0}, 21)
	if got.Page != 1 || got.PageSize != 10 || got.TotalPages != 3 {
		t.Fatalf("page info = %+v", got)
	}
}

func TestTokenView_preservesOptionalTimesWithoutRawToken(t *testing.T) {
	expires := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	view := tokenView(&ent.ApiToken{ID: "token-1", ExpiresAt: expires, LastUsedAt: expires})
	if view.ExpiresAt == nil || !view.ExpiresAt.Equal(expires) || view.LastUsedAt == nil || !view.LastUsedAt.Equal(expires) {
		t.Fatalf("optional token dates = %+v", view)
	}
}

func TestWriteAuditCSV_quotesCommasAndNewlines(t *testing.T) {
	entity := &ent.AuditLog{CreatedAt: time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC), Action: "create, user", Resource: "users", Details: "line 1\nline 2", Level: "info", Status: "success"}
	var output bytes.Buffer
	if err := writeAuditCSV(&output, []*ent.AuditLog{entity}); err != nil {
		t.Fatalf("write CSV: %v", err)
	}
	if !strings.Contains(output.String(), `"create, user"`) || !strings.Contains(output.String(), "\"line 1\nline 2\"") {
		t.Fatalf("CSV did not quote fields: %q", output.String())
	}
}
