-- CreateTable
CREATE TABLE "McpService" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "identifier" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "command" TEXT,
    "endpoint" TEXT,
    "envVars" TEXT,
    "pricing" INTEGER NOT NULL DEFAULT 0,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "callCount" INTEGER NOT NULL DEFAULT 0,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "McpUsage" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "userId" TEXT NOT NULL,
    "mcpId" TEXT NOT NULL,
    "credits" INTEGER NOT NULL,
    "status" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "McpUsage_userId_fkey" FOREIGN KEY ("userId") REFERENCES "User" ("id") ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT "McpUsage_mcpId_fkey" FOREIGN KEY ("mcpId") REFERENCES "McpService" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateIndex
CREATE UNIQUE INDEX "McpService_identifier_key" ON "McpService"("identifier");

-- CreateIndex
CREATE INDEX "McpService_type_idx" ON "McpService"("type");

-- CreateIndex
CREATE INDEX "McpService_isActive_idx" ON "McpService"("isActive");

-- CreateIndex
CREATE INDEX "McpUsage_userId_idx" ON "McpUsage"("userId");

-- CreateIndex
CREATE INDEX "McpUsage_mcpId_idx" ON "McpUsage"("mcpId");
