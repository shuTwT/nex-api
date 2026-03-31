/*
  Warnings:

  - You are about to drop the column `planId` on the `Payment` table. All the data in the column will be lost.

*/
-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_Payment" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "userId" TEXT NOT NULL,
    "outTradeNo" TEXT NOT NULL,
    "transactionId" TEXT,
    "method" TEXT NOT NULL,
    "amount" REAL NOT NULL,
    "currency" TEXT NOT NULL DEFAULT 'CNY',
    "status" TEXT NOT NULL DEFAULT 'pending',
    "qrcodeUrl" TEXT,
    "payUrl" TEXT,
    "notifyUrl" TEXT,
    "paidAt" DATETIME,
    "expiredAt" DATETIME,
    "cancelledAt" DATETIME,
    "metadata" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "Payment_userId_fkey" FOREIGN KEY ("userId") REFERENCES "User" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);
INSERT INTO "new_Payment" ("amount", "cancelledAt", "createdAt", "currency", "expiredAt", "id", "metadata", "method", "notifyUrl", "outTradeNo", "paidAt", "payUrl", "qrcodeUrl", "status", "transactionId", "updatedAt", "userId") SELECT "amount", "cancelledAt", "createdAt", "currency", "expiredAt", "id", "metadata", "method", "notifyUrl", "outTradeNo", "paidAt", "payUrl", "qrcodeUrl", "status", "transactionId", "updatedAt", "userId" FROM "Payment";
DROP TABLE "Payment";
ALTER TABLE "new_Payment" RENAME TO "Payment";
CREATE UNIQUE INDEX "Payment_outTradeNo_key" ON "Payment"("outTradeNo");
CREATE INDEX "Payment_userId_idx" ON "Payment"("userId");
CREATE INDEX "Payment_status_idx" ON "Payment"("status");
CREATE INDEX "Payment_method_idx" ON "Payment"("method");
CREATE INDEX "Payment_outTradeNo_idx" ON "Payment"("outTradeNo");
CREATE INDEX "Payment_transactionId_idx" ON "Payment"("transactionId");
CREATE INDEX "Payment_createdAt_idx" ON "Payment"("createdAt");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
