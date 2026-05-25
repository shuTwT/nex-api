import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function POST(request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const body = await request.json();
    const { code: codeInput } = body;

    const trimmed = codeInput?.trim().toUpperCase();
    if (!trimmed) return apiError("请输入兑换码", 400);

    const code = await prisma.redemptionCode.findUnique({ where: { code: trimmed } });
    if (!code) return apiError("兑换码不存在", 404);
    if (code.isUsed) return apiError("该兑换码已被使用", 400);
    if (code.expiresAt && new Date(code.expiresAt) < new Date()) return apiError("该兑换码已过期", 400);

    if (code.type === "quota") {
      const credits = code.credits || 0;
      await prisma.user.update({ where: { id: sessionUser.id }, data: { credits: { increment: credits } } });
    } else if (code.type === "subscription") {
      const plan = await prisma.subscriptionPlan.findUnique({ where: { id: code.planId || "" } });
      if (!plan) return apiError("关联的订阅计划不存在", 404);

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

      await prisma.subscription.create({
        data: { userId: sessionUser.id, planId: plan.id, planName: plan.title, credits: plan.totalCredits, price: 0, startDate: now, endDate, isActive: true },
      });
    }

    await prisma.redemptionCode.update({
      where: { id: code.id },
      data: { isUsed: true, usedBy: sessionUser.id, usedAt: new Date() },
    });

    const message = code.type === "quota"
      ? `兑换成功！获得 ${(code.credits || 0).toLocaleString()} 额度`
      : `兑换成功！获得 ${code.planName || ""} 订阅`;

    return apiSuccess({ message });
  } catch (error) {
    console.error("Error redeeming code:", error);
    return apiError("兑换失败，请重试");
  }
}
