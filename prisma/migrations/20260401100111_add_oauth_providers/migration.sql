-- CreateTable
CREATE TABLE "OAuthProvider" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "clientId" TEXT NOT NULL,
    "clientSecret" TEXT NOT NULL,
    "authorizationUrl" TEXT NOT NULL,
    "tokenUrl" TEXT NOT NULL,
    "userInfoUrl" TEXT NOT NULL,
    "scopes" TEXT NOT NULL,
    "userIdField" TEXT NOT NULL,
    "emailField" TEXT NOT NULL,
    "usernameField" TEXT NOT NULL,
    "roleField" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateIndex
CREATE UNIQUE INDEX "OAuthProvider_name_key" ON "OAuthProvider"("name");

-- CreateIndex
CREATE INDEX "OAuthProvider_name_idx" ON "OAuthProvider"("name");

-- CreateIndex
CREATE INDEX "OAuthProvider_isActive_idx" ON "OAuthProvider"("isActive");
