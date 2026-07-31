package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
)

const catalogSchema = `
CREATE TABLE "ApiCategory" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL UNIQUE, "description" TEXT NOT NULL, "icon" TEXT);
CREATE TABLE "Api" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL, "alias" TEXT NOT NULL UNIQUE, "description" TEXT NOT NULL, "endpoint" TEXT NOT NULL UNIQUE, "method" TEXT NOT NULL, "categoryId" TEXT NOT NULL, "pricing" INTEGER NOT NULL DEFAULT 0, "documentation" TEXT, "preScript" TEXT, "postScript" TEXT, "isActive" BOOLEAN NOT NULL DEFAULT 1, "callCount" INTEGER NOT NULL DEFAULT 0, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, "updatedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY ("categoryId") REFERENCES "ApiCategory"("id") ON DELETE RESTRICT);
CREATE TABLE "ApiParameter" ("id" TEXT PRIMARY KEY NOT NULL, "apiId" TEXT NOT NULL, "name" TEXT NOT NULL, "type" TEXT NOT NULL, "required" BOOLEAN NOT NULL, "description" TEXT NOT NULL, "defaultValue" TEXT, FOREIGN KEY ("apiId") REFERENCES "Api"("id") ON DELETE RESTRICT);
CREATE TABLE "ApiResponse" ("id" TEXT PRIMARY KEY NOT NULL, "apiId" TEXT NOT NULL, "name" TEXT NOT NULL, "type" TEXT NOT NULL, "description" TEXT NOT NULL, FOREIGN KEY ("apiId") REFERENCES "Api"("id") ON DELETE RESTRICT);
CREATE TABLE "ApiUsage" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT NOT NULL, "apiId" TEXT NOT NULL, "credits" INTEGER NOT NULL, "status" TEXT NOT NULL, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY ("apiId") REFERENCES "Api"("id") ON DELETE RESTRICT);
CREATE TABLE "McpService" ("id" TEXT PRIMARY KEY NOT NULL, "name" TEXT NOT NULL, "identifier" TEXT NOT NULL UNIQUE, "type" TEXT NOT NULL, "command" TEXT, "endpoint" TEXT, "envVars" TEXT, "pricing" INTEGER NOT NULL DEFAULT 0, "isActive" BOOLEAN NOT NULL DEFAULT 1, "callCount" INTEGER NOT NULL DEFAULT 0, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, "updatedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE "McpUsage" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT NOT NULL, "mcpId" TEXT NOT NULL, "credits" INTEGER NOT NULL, "status" TEXT NOT NULL, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY ("mcpId") REFERENCES "McpService"("id") ON DELETE RESTRICT);
CREATE TABLE "AuditLog" ("id" TEXT PRIMARY KEY NOT NULL, "userId" TEXT, "action" TEXT NOT NULL, "resource" TEXT NOT NULL, "details" TEXT, "ipAddress" TEXT, "userAgent" TEXT, "level" TEXT NOT NULL DEFAULT 'info', "status" TEXT NOT NULL DEFAULT 'success', "metadata" TEXT, "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
`

func TestCatalogValidation_rejectsInvalidIdentifiersAndMethods(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{name: "alias", call: func() error { return validateAlias("1bad") }},
		{name: "method", call: func() error { return validateMethod("TRACE") }},
		{name: "mcp identifier", call: func() error { return validateIdentifier("-bad") }},
		{name: "mcp type", call: func() error { return validateMCPType("websocket") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			err := test.call()

			// Then
			if err == nil || !strings.Contains(err.Error(), "validation") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestHandler_rejectsUnknownJSONFields(t *testing.T) {
	// Given
	client := newCatalogClient(t)
	apis, err := NewAPIService(client)
	if err != nil {
		t.Fatal(err)
	}
	categories, err := NewCategoryService(client)
	if err != nil {
		t.Fatal(err)
	}
	mcp, err := NewMCPService(client)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(apis, categories, mcp)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	if err := RegisterRoutes(mux, handler); err != nil {
		t.Fatal(err)
	}

	// When
	request := httptest.NewRequest(http.MethodPost, "/api/categories", strings.NewReader(`{"name":"tools","unexpected":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	// Then
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestCatalogServices_CRUDStatsToggleAndAudit(t *testing.T) {
	// Given
	client := newCatalogClient(t)
	apis, _ := NewAPIService(client)
	categories, _ := NewCategoryService(client)
	mcp, _ := NewMCPService(client)
	ctx := context.Background()
	category, err := categories.Create(ctx, CategoryInput{Name: "Tools", Description: "utility"})
	if err != nil {
		t.Fatal(err)
	}

	// When: create and list an API
	item, err := apis.Create(ctx, APIInput{Name: "Weather", Alias: "weather", Description: "forecast", Endpoint: "/weather", Method: "get", CategoryID: category.ID, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = apis.Create(ctx, APIInput{Name: "Duplicate", Alias: "weather", Endpoint: "/other", Method: "GET", CategoryID: category.ID, IsActive: true})
	if err == nil || !errors.Is(err, runtime.ErrConflict) {
		t.Fatalf("duplicate API alias error = %v", err)
	}
	_, err = categories.Create(ctx, CategoryInput{Name: "Tools"})
	if err == nil || !errors.Is(err, runtime.ErrConflict) {
		t.Fatalf("duplicate category name error = %v", err)
	}
	result, err := apis.List(ctx, APIListOptions{Page: 1, Limit: 10})
	if err != nil || result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("list result = %+v, err = %v", result, err)
	}

	// Then: update, toggle, stats, and audit are observable
	updated, err := apis.Update(ctx, item.ID, APIUpdateInput{Description: new("updated")})
	if err != nil || updated.Description != "updated" {
		t.Fatalf("updated = %+v, err = %v", updated, err)
	}
	updated, err = apis.Toggle(ctx, item.ID)
	if err != nil || updated.IsActive {
		t.Fatalf("toggle = %+v, err = %v", updated, err)
	}
	stats, err := apis.Stats(ctx)
	if err != nil || stats.TotalAPIs != 1 || stats.InactiveAPIs != 1 || stats.CategoriesCount != 1 {
		t.Fatalf("stats = %+v, err = %v", stats, err)
	}

	mcpItem, err := mcp.Create(ctx, MCPInput{Name: "Browser", Identifier: "browser", Type: "sse", Endpoint: new("mcp.example"), IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = mcp.Create(ctx, MCPInput{Name: "Duplicate", Identifier: "browser", Type: "sse", Endpoint: new("duplicate"), IsActive: true})
	if err == nil || !errors.Is(err, runtime.ErrConflict) {
		t.Fatalf("duplicate MCP identifier error = %v", err)
	}
	mcpItem, err = mcp.Toggle(ctx, mcpItem.ID)
	if err != nil || mcpItem.IsActive {
		t.Fatalf("MCP toggle = %+v, err = %v", mcpItem, err)
	}
	mcpStats, err := mcp.Stats(ctx)
	if err != nil || mcpStats.TotalServices != 1 || mcpStats.InactiveServices != 1 {
		t.Fatalf("MCP stats = %+v, err = %v", mcpStats, err)
	}
	auditCount, err := client.AuditLog.Query().Count(ctx)
	if err != nil || auditCount < 6 {
		t.Fatalf("audit count = %d, err = %v", auditCount, err)
	}
}

func TestCatalogDeletes_returnConflictForLegacyDependents(t *testing.T) {
	// Given
	database, client := newCatalogDatabase(t)
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
	if _, err := database.ExecContext(ctx, `INSERT INTO ApiParameter (id, apiId, name, type, required, description) VALUES ('param-1', ?, 'city', 'string', 1, '')`, item.ID); err != nil {
		t.Fatal(err)
	}
	service, err := mcp.Create(ctx, MCPInput{Name: "Browser", Identifier: "browser", Type: "stdio", Command: new("browser"), IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `INSERT INTO McpUsage (id, userId, mcpId, credits, status) VALUES ('usage-1', 'user-1', ?, 1, 'success')`, service.ID); err != nil {
		t.Fatal(err)
	}

	// When
	apiDeleteErr := apis.Delete(ctx, item.ID)
	categoryDeleteErr := categories.Delete(ctx, category.ID)
	mcpDeleteErr := mcp.Delete(ctx, service.ID)

	// Then
	for name, deleteErr := range map[string]error{"api": apiDeleteErr, "category": categoryDeleteErr, "mcp": mcpDeleteErr} {
		if deleteErr == nil || !errors.Is(deleteErr, runtime.ErrConflict) {
			t.Errorf("%s delete error = %v", name, deleteErr)
		}
	}
}

func newCatalogClient(t *testing.T) *ent.Client {
	_, client := newCatalogDatabase(t)
	return client
}

func newCatalogDatabase(t *testing.T) (*sql.DB, *ent.Client) {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "catalog.db")
	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.ExecContext(context.Background(), catalogSchema); err != nil {
		t.Fatalf("create catalog schema: %v", err)
	}
	client, err := ent.Open(dialect.SQLite, databasePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return database, client
}

func TestCatalogResponseShape_includesEnvelope(t *testing.T) {
	// Given
	client := newCatalogClient(t)
	apis, _ := NewAPIService(client)
	categories, _ := NewCategoryService(client)
	mcp, _ := NewMCPService(client)
	handler, _ := NewHandler(apis, categories, mcp)
	mux := http.NewServeMux()
	_ = RegisterRoutes(mux, handler)

	// When
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/categories", nil))

	// Then
	var body struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || string(body.Data) != "[]" {
		t.Fatalf("body = %s", body.Data)
	}
}
