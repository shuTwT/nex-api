import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import type { BusinessCallbackPayload } from "@/lib/payment/callback";

export async function POST(request: NextRequest) {
  try {
    const payload: BusinessCallbackPayload = await request.json();

    console.log("收到订阅业务回调:", payload);

    if (payload.status !== "paid") {
      return NextResponse.json({ 
        success: true, 
        message: "非支付成功状态，跳过处理" 
      });
    }

    const planId = payload.metadata?.planId as string | undefined;
    if (!planId) {
      return NextResponse.json({ 
        success: true, 
        message: "无订阅计划ID，跳过处理" 
      });
    }

    const plan = await prisma.subscriptionPlan.findUnique({
      where: { id: planId },
    });

    if (!plan) {
      return NextResponse.json(
        { success: false, error: "订阅计划不存在" },
        { status: 400 }
      );
    }

    await prisma.$transaction(async (tx) => {
      const existingSubscription = await tx.subscription.findFirst({
        where: {
          userId: payload.userId,
          isActive: true,
          endDate: {
            gte: new Date(),
          },
        },
      });

      if (existingSubscription) {
        await tx.subscription.update({
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

      await tx.subscription.create({
        data: {
          userId: payload.userId,
          planId,
          planName: plan.title,
          credits: plan.totalCredits,
          price: payload.amount,
          startDate: now,
          endDate: endDate,
          isActive: true,
          paymentId: payload.paymentId,
        },
      });

      await tx.user.update({
        where: { id: payload.userId },
        data: {
          credits: {
            increment: plan.totalCredits,
          },
        },
      });
    });

    console.log("订阅创建成功");

    return NextResponse.json({ 
      success: true, 
      message: "订阅创建成功" 
    });
  } catch (error) {
    console.error("处理订阅业务回调失败:", error);
    return NextResponse.json(
      { 
        success: false, 
        error: error instanceof Error ? error.message : "处理失败" 
      },
      { status: 500 }
    );
  }
}
