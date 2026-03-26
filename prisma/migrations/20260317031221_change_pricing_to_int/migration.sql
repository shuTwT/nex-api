/*
  Warnings:

  - You are about to alter the column `pricing` on the `Api` table. The data in that column could be lost. The data in that column will be cast from `String` to `Int`.

*/
-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_Api" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "alias" TEXT,
    "description" TEXT NOT NULL,
    "endpoint" TEXT NOT NULL,
    "method" TEXT NOT NULL,
    "categoryId" TEXT NOT NULL,
    "pricing" INTEGER NOT NULL DEFAULT 0,
    "documentation" TEXT,
    "preScript" TEXT,
    "postScript" TEXT,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "callCount" INTEGER NOT NULL DEFAULT 0,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "Api_categoryId_fkey" FOREIGN KEY ("categoryId") REFERENCES "ApiCategory" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);
INSERT INTO "new_Api" ("alias", "callCount", "categoryId", "createdAt", "description", "documentation", "endpoint", "id", "isActive", "method", "name", "postScript", "preScript", "pricing", "updatedAt") SELECT "alias", "callCount", "categoryId", "createdAt", "description", "documentation", "endpoint", "id", "isActive", "method", "name", "postScript", "preScript", "pricing", "updatedAt" FROM "Api";
DROP TABLE "Api";
ALTER TABLE "new_Api" RENAME TO "Api";
CREATE UNIQUE INDEX "Api_alias_key" ON "Api"("alias");
CREATE UNIQUE INDEX "Api_endpoint_key" ON "Api"("endpoint");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
