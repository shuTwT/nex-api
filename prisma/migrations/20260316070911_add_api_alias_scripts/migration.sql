/*
  Warnings:

  - A unique constraint covering the columns `[alias]` on the table `Api` will be added. If there are existing duplicate values, this will fail.

*/
-- AlterTable
ALTER TABLE "Api" ADD COLUMN "alias" TEXT;
ALTER TABLE "Api" ADD COLUMN "postScript" TEXT;
ALTER TABLE "Api" ADD COLUMN "preScript" TEXT;

-- CreateIndex
CREATE UNIQUE INDEX "Api_alias_key" ON "Api"("alias");
