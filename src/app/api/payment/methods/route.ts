import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { PaymentServiceFactory, type PaymentMethod } from "@/lib/payment";
import { getAvailablePaymentMethods } from "@/lib/payment";

export async function GET(_request: NextRequest) {
  try {
    const methods = await getAvailablePaymentMethods();
    return apiSuccess(methods);
  } catch (error) {
    console.error("Error fetching payment methods:", error);
    return apiError("获取支付方式失败");
  }
}

export async function POST(request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const body = await request.json();
    const { planId, method } = body;

    const plan = await prisma.subscriptionPlan.findUnique({ where: { id: planId } });
    if (!plan) return apiError("订阅计划不存在", 404);
    if (!plan.isActive) return apiError("该订阅计划已下线", 400);

    const notifyUrl = `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/business/subscription`;

    const result = await PaymentServiceFactory.createPayment({
      userId: sessionUser.id, amount: plan.price, method: method as PaymentMethod, notifyUrl,
      metadata: { type: "subscription", planId: plan.id },
    });

    if (!result.success) return apiError(result.error || "创建支付失败");

    revalidatePath("/payment");
    return apiSuccess(result, 201);
  } catch (error) {
    console.error("Error creating payment:", error);
    return apiError("创建支付失败");
  }
}
