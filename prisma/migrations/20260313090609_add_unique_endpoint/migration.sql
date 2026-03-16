/*
  Warnings:

  - A unique constraint covering the columns `[endpoint]` on the table `Api` will be added. If there are existing duplicate values, this will fail.

*/
-- CreateIndex
CREATE UNIQUE INDEX "Api_endpoint_key" ON "Api"("endpoint");
