"use server";

import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/auth/util";
import { authOptions } from "@/lib/auth/config";

export async function getCurrentUserProfile() {
  try {
    const sessionUser = await requireAuth(authOptions);

    console.log(sessionUser)
    
    const user = await prisma.user.findUnique({
      where: { id: sessionUser.id },
      select: {
        id: true,
        name: true,
        email: true,
        image: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
      },
    });

    if (!user) {
      return { success: false, error: "用户未找到" };
    }

    const totalCreditsSpent = await prisma.apiUsage.aggregate({
      where: {
        userId: user.id,
      },
      _sum: {
        credits: true,
      },
    });

    const totalRequests = await prisma.apiUsage.count({
      where: {
        userId: user.id,
      },
    });

    return {
      success: true,
      data: {
        ...user,
        totalCreditsSpent: totalCreditsSpent._sum.credits || 0,
        totalRequests,
      },
    };
  } catch (error) {
    console.error("Error fetching user profile:", error);
    return { success: false, error: "获取用户信息失败" };
  }
}

export async function redeemCode(codeInput: string) {
  try {
    const sessionUser = await requireAuth(authOptions);

    const trimmed = codeInput.trim().toUpperCase();
    if (!trimmed) {
      return { success: false, error: "请输入兑换码" };
    }

    const code = await prisma.redemptionCode.findUnique({
      where: { code: trimmed },
    });

    if (!code) {
      return { success: false, error: "兑换码不存在" };
    }

    if (code.isUsed) {
      return { success: false, error: "该兑换码已被使用" };
    }

    if (code.expiresAt && new Date(code.expiresAt) < new Date()) {
      return { success: false, error: "该兑换码已过期" };
    }

    if (code.type === "quota") {
      const credits = code.credits || 0;
      await prisma.user.update({
        where: { id: sessionUser.id },
        data: { credits: { increment: credits } },
      });
    } else if (code.type === "subscription") {
      const plan = await prisma.subscriptionPlan.findUnique({
        where: { id: code.planId || "" },
      });

      if (!plan) {
        return { success: false, error: "关联的订阅计划不存在" };
      }

      const existingSubscription = await prisma.subscription.findFirst({
        where: {
          userId: sessionUser.id,
          isActive: true,
          endDate: { gte: new Date() },
        },
      });

      if (existingSubscription) {
        await prisma.subscription.update({
          where: { id: existingSubscription.id },
          data: { isActive: false },
        });
      }

      const now = new Date();
      const endDate = new Date(now);
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

      await prisma.subscription.create({
        data: {
          userId: sessionUser.id,
          planId: plan.id,
          planName: plan.title,
          credits: plan.totalCredits,
          price: 0,
          startDate: now,
          endDate,
          isActive: true,
        },
      });
    }

    await prisma.redemptionCode.update({
      where: { id: code.id },
      data: {
        isUsed: true,
        usedBy: sessionUser.id,
        usedAt: new Date(),
      },
    });

    return {
      success: true,
      message:
        code.type === "quota"
          ? `兑换成功！获得 ${(code.credits || 0).toLocaleString()} 额度`
          : `兑换成功！获得 ${code.planName || ""} 订阅`,
    };
  } catch (error) {
    console.error("Error redeeming code:", error);
    return { success: false, error: "兑换失败，请重试" };
  }
}
