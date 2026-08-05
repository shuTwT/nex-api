package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/enttest"
	"github.com/shuTwT/nex-api/internal/service/apierror"
)

func TestCatalogValidation_rejectsInvalidIdentifiersAndMethods(t *testing.T) {
	for _, test := range []struct {
		name string
		call func() error
	}{
		{"alias", func() error { return validateAlias("1bad") }}, {"method", func() error { return validateMethod("TRACE") }}, {"mcp identifier", func() error { return validateIdentifier("-bad") }}, {"mcp type", func() error { return validateMCPType("websocket") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil || !strings.Contains(err.Error(), "validation") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestCatalogServices_CRUDStatsToggleAndAudit(t *testing.T) {
	client := newCatalogClient(t)
	apis, _ := NewAPIService(client)
	categories, _ := NewCategoryService(client)
	mcp, _ := NewMCPService(client)
	ctx := context.Background()
	category, err := categories.Create(ctx, CategoryInput{Name: "Tools", Description: "utility"})
	if err != nil {
		t.Fatal(err)
	}
	item, err := apis.Create(ctx, APIInput{Name: "Weather", Alias: "weather", Description: "forecast", Endpoint: "/weather", Method: "get", CategoryID: category.ID, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := apis.Create(ctx, APIInput{Name: "Duplicate", Alias: "weather", Endpoint: "/other", Method: "GET", CategoryID: category.ID, IsActive: true}); err == nil || !errors.Is(err, apierror.ErrConflict) {
		t.Fatalf("duplicate API error = %v", err)
	}
	if _, err := categories.Create(ctx, CategoryInput{Name: "Tools"}); err == nil || !errors.Is(err, apierror.ErrConflict) {
		t.Fatalf("duplicate category error = %v", err)
	}
	result, err := apis.List(ctx, APIListOptions{Page: 1, Limit: 10})
	if err != nil || result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("list = %+v, %v", result, err)
	}
	updated, err := apis.Update(ctx, item.ID, APIUpdateInput{Description: stringPtr("updated")})
	if err != nil || updated.Description != "updated" {
		t.Fatalf("updated = %+v, %v", updated, err)
	}
	updated, err = apis.Toggle(ctx, item.ID)
	if err != nil || updated.IsActive {
		t.Fatalf("toggle = %+v, %v", updated, err)
	}
	stats, err := apis.Stats(ctx)
	if err != nil || stats.TotalAPIs != 1 || stats.InactiveAPIs != 1 || stats.CategoriesCount != 1 {
		t.Fatalf("stats = %+v, %v", stats, err)
	}
	mcpItem, err := mcp.Create(ctx, MCPInput{Name: "Browser", Identifier: "browser", CategoryID: category.ID, Description: stringPtr("浏览器自动化服务"), Documentation: stringPtr("https://example.com/mcp-docs"), Type: "sse", Endpoint: stringPtr("mcp.example"), IsActive: true})
	if err != nil || mcpItem.CategoryId != category.ID || mcpItem.Description != "浏览器自动化服务" || mcpItem.Documentation != "https://example.com/mcp-docs" {
		t.Fatalf("MCP create = %+v, %v", mcpItem, err)
	}
	if _, err := mcp.Create(ctx, MCPInput{Name: "Duplicate", Identifier: "browser", CategoryID: category.ID, Type: "sse", Endpoint: stringPtr("duplicate"), IsActive: true}); err == nil || !errors.Is(err, apierror.ErrConflict) {
		t.Fatalf("duplicate MCP error = %v", err)
	}
	mcpItem, err = mcp.Toggle(ctx, mcpItem.ID)
	if err != nil || mcpItem.IsActive {
		t.Fatalf("MCP toggle = %+v, %v", mcpItem, err)
	}
	mcpStats, err := mcp.Stats(ctx)
	if err != nil || mcpStats.TotalServices != 1 || mcpStats.InactiveServices != 1 {
		t.Fatalf("MCP stats = %+v, %v", mcpStats, err)
	}
	audits, err := client.AuditLog.Query().Count(ctx)
	if err != nil || audits < 6 {
		t.Fatalf("audits = %d, %v", audits, err)
	}
}

func TestCatalogDeletes_returnConflictForLegacyDependents(t *testing.T) {
	client := newCatalogClient(t)
	apis, _ := NewAPIService(client)
	categories, _ := NewCategoryService(client)
	mcp, _ := NewMCPService(client)
	ctx := context.Background()
	category, err := categories.Create(ctx, CategoryInput{Name: "Tools"})
	if err != nil {
		t.Fatal(err)
	}
	item, err := apis.Create(ctx, APIInput{Name: "Weather", Alias: "weather", Endpoint: "/weather", Method: "GET", CategoryID: category.ID, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ApiParameter.Create().SetID("param-1").SetApiId(item.ID).SetName("city").SetType("string").SetRequired(true).SetDescription("").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.User.Create().SetID("user-1").SetEmail("user@example.com").SetUsername("user").SetPassword("redacted").SetRole("user").SetCredits(0).Save(ctx); err != nil {
		t.Fatal(err)
	}
	service, err := mcp.Create(ctx, MCPInput{Name: "Browser", Identifier: "browser", CategoryID: category.ID, Type: "stdio", Command: stringPtr("browser"), IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.McpUsage.Create().SetID("usage-1").SetUserId("user-1").SetMcpId(service.ID).SetCredits(1).SetStatus("success").Save(ctx); err != nil {
		t.Fatal(err)
	}
	for name, err := range map[string]error{"api": apis.Delete(ctx, item.ID), "category": categories.Delete(ctx, category.ID), "mcp": mcp.Delete(ctx, service.ID)} {
		if err == nil || !errors.Is(err, apierror.ErrConflict) {
			t.Errorf("%s delete error = %v", name, err)
		}
	}
}

func newCatalogClient(t *testing.T) *ent.Client {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:"+t.TempDir()+"/catalog.db?_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	return client
}
func stringPtr(value string) *string { return &value }
