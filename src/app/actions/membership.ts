"use server";

import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/auth/util";
import { authOptions } from "@/lib/auth/config";

export async function getCurrentSubscription() {
  try {
    const sessionUser = await requireAuth(authOptions);

    const subscription = await prisma.subscription.findFirst({
      where: {
        userId: sessionUser.id,
        isActive: true,
        endDate: {
          gte: new Date(),
        },
      },
      include: {
        plan: true,
      },
      orderBy: {
        createdAt: "desc",
      },
    });

    return { success: true, data: subscription };
  } catch (error) {
    console.error("Error fetching current subscription:", error);
    return { success: false, error: "获取订阅信息失败" };
  }
}

export async function getAvailablePlans() {
  try {
    const plans = await prisma.subscriptionPlan.findMany({
      where: {
        isActive: true,
      },
      orderBy: {
        sortOrder: "asc",
      },
    });

    return { success: true, data: plans };
  } catch (error) {
    console.error("Error fetching available plans:", error);
    return { success: false, error: "获取订阅计划失败" };
  }
}

export async function subscribeToPlan(planId: string) {
  try {
    const sessionUser = await requireAuth(authOptions);

    const plan = await prisma.subscriptionPlan.findUnique({
      where: { id: planId },
    });

    if (!plan) {
      return { success: false, error: "订阅计划不存在" };
    }

    if (!plan.isActive) {
      return { success: false, error: "该订阅计划已下线" };
    }

    const existingSubscription = await prisma.subscription.findFirst({
      where: {
        userId: sessionUser.id,
        isActive: true,
        endDate: {
          gte: new Date(),
        },
      },
    });

    if (existingSubscription) {
      await prisma.subscription.update({
        where: { id: existingSubscription.id },
        data: { isActive: false },
      });
    }

    const now = new Date();
    let endDate = new Date(now);

    switch (plan.validityUnit) {
      case "day":
        endDate.setDate(endDate.getDate() + plan.validityDuration);
        break;
      case "week":
        endDate.setDate(endDate.getDate() + plan.validityDuration * 7);
        break;
      case "month":
        endDate.setMonth(endDate.getMonth() + plan.validityDuration);
        break;
      case "year":
        endDate.setFullYear(endDate.getFullYear() + plan.validityDuration);
        break;
    }

    const subscription = await prisma.subscription.create({
      data: {
        userId: sessionUser.id,
        planId: plan.id,
        planName: plan.title,
        credits: plan.totalCredits,
        price: plan.price,
        startDate: now,
        endDate: endDate,
        isActive: true,
      },
    });

    await prisma.user.update({
      where: { id: sessionUser.id },
      data: {
        credits: {
          increment: plan.totalCredits,
        },
      },
    });

    return { success: true, data: subscription };
  } catch (error) {
    console.error("Error subscribing to plan:", error);
    return { success: false, error: "订阅失败" };
  }
}
