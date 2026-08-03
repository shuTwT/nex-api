CREATE TABLE "ScheduledJob" (
  "id" TEXT NOT NULL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "taskKey" TEXT NOT NULL,
  "scheduleType" TEXT NOT NULL,
  "expression" TEXT NOT NULL,
  "enabled" BOOLEAN NOT NULL DEFAULT true,
  "description" TEXT,
  "lastRunAt" DATETIME,
  "lastStatus" TEXT NOT NULL DEFAULT 'never',
  "lastError" TEXT,
  "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" DATETIME NOT NULL
);

CREATE UNIQUE INDEX "ScheduledJob_taskKey_key" ON "ScheduledJob"("taskKey");
CREATE INDEX "ScheduledJob_enabled_idx" ON "ScheduledJob"("enabled");
