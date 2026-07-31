ALTER TABLE "Payment" ADD COLUMN "callbackVersion" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "Payment" ADD COLUMN "callbackKey" TEXT;
ALTER TABLE "Payment" ADD COLUMN "callbackProcessedAt" DATETIME;

CREATE UNIQUE INDEX "Payment_callbackKey_key" ON "Payment"("callbackKey");
CREATE INDEX "Payment_callbackVersion_idx" ON "Payment"("callbackVersion");
