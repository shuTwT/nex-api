import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function POST(request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const body = await request.json();
    const { planId } = body;

    const plan = await prisma.subscriptionPlan.findUnique({ where: { id: planId } });
    if (!plan) return apiError("订阅计划不存在", 404);
    if (!plan.isActive) return apiError("该订阅计划已下线", 400);

    const existingSubscription = await prisma.subscription.findFirst({
      where: { userId: sessionUser.id, isActive: true, endDate: { gte: new Date() } },
    });
    if (existingSubscription) {
      await prisma.subscription.update({ where: { id: existingSubscription.id }, data: { isActive: false } });
    }

    const now = new Date();
    const endDate = new Date(now);
    switch (plan.validityUnit) {
      case "day": endDate.setDate(endDate.getDate() + plan.validityDuration); break;
      case "week": endDate.setDate(endDate.getDate() + plan.validityDuration * 7); break;
      case "month": endDate.setMonth(endDate.getMonth() + plan.validityDuration); break;
      case "year": endDate.setFullYear(endDate.getFullYear() + plan.validityDuration); break;
    }

    const subscription = await prisma.subscription.create({
      data: { userId: sessionUser.id, planId: plan.id, planName: plan.title, credits: plan.totalCredits, price: plan.price, startDate: now, endDate, isActive: true },
    });

    await prisma.user.update({ where: { id: sessionUser.id }, data: { credits: { increment: plan.totalCredits } } });

    return apiSuccess(subscription, 201);
  } catch (error) {
    console.error("Error subscribing to plan:", error);
    return apiError("订阅失败");
  }
}
