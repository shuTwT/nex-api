-- CreateTable
CREATE TABLE "RedemptionCode" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "code" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "planId" TEXT,
    "planName" TEXT,
    "credits" INTEGER,
    "expiresAt" DATETIME,
    "isUsed" BOOLEAN NOT NULL DEFAULT false,
    "usedBy" TEXT,
    "usedAt" DATETIME,
    "createdBy" TEXT NOT NULL,
    "batchId" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateIndex
CREATE UNIQUE INDEX "RedemptionCode_code_key" ON "RedemptionCode"("code");

-- CreateIndex
CREATE INDEX "RedemptionCode_code_idx" ON "RedemptionCode"("code");

-- CreateIndex
CREATE INDEX "RedemptionCode_type_idx" ON "RedemptionCode"("type");

-- CreateIndex
CREATE INDEX "RedemptionCode_isUsed_idx" ON "RedemptionCode"("isUsed");

-- CreateIndex
CREATE INDEX "RedemptionCode_batchId_idx" ON "RedemptionCode"("batchId");

-- CreateIndex
CREATE INDEX "RedemptionCode_createdBy_idx" ON "RedemptionCode"("createdBy");
