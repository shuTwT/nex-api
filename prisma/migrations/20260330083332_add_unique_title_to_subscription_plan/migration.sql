/*
  Warnings:

  - A unique constraint covering the columns `[title]` on the table `SubscriptionPlan` will be added. If there are existing duplicate values, this will fail.

*/
-- CreateIndex
CREATE UNIQUE INDEX "SubscriptionPlan_title_key" ON "SubscriptionPlan"("title");
