-- Add MCP marketplace metadata and link MCP services to the shared API category catalog.
INSERT INTO "ApiCategory" ("id", "name", "description", "icon")
SELECT 'mcp-category-uncategorized', '未分类', '尚未归类的服务', NULL
WHERE NOT EXISTS (
  SELECT 1 FROM "ApiCategory" WHERE "name" = '未分类'
);

ALTER TABLE "McpService" ADD COLUMN "description" TEXT;
ALTER TABLE "McpService" ADD COLUMN "documentation" TEXT;
ALTER TABLE "McpService" ADD COLUMN "categoryId" TEXT REFERENCES "ApiCategory" ("id") ON DELETE RESTRICT ON UPDATE NO ACTION;

UPDATE "McpService"
SET "categoryId" = (
  SELECT "id" FROM "ApiCategory" WHERE "name" = '未分类' LIMIT 1
)
WHERE "categoryId" IS NULL;

CREATE INDEX "McpService_categoryId_idx" ON "McpService" ("categoryId");
