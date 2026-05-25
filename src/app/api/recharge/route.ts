import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { PaymentServiceFactory, type PaymentMethod } from "@/lib/payment";

export async function POST(request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const body = await request.json();
    const { amount, credits, method } = body;

    const settings = await prisma.systemSetting.findMany({ where: { category: "payment" } });
    const settingsMap: Record<string, string> = {};
    settings.forEach((s) => { settingsMap[s.key] = s.value; });

    const minRecharge = parseFloat(settingsMap.minRecharge || "10");
    const creditPrice = parseFloat(settingsMap.creditPrice || "1");

    if (amount < minRecharge) return apiError(`最低充值金额为 ¥${minRecharge}`, 400);

    const expectedCredits = Math.floor(amount / creditPrice);
    if (credits !== expectedCredits) return apiError("积分计算错误", 400);

    const notifyUrl = `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/business/recharge`;

    const result = await PaymentServiceFactory.createPayment({
      userId: sessionUser.id, planId: undefined, amount, method: method as PaymentMethod, notifyUrl,
      metadata: { type: "recharge", credits, creditPrice },
    });

    if (!result.success) return apiError(result.error || "创建支付失败");

    revalidatePath("/payment");
    return apiSuccess(result, 201);
  } catch (error) {
    console.error("Error creating recharge payment:", error);
    return apiError("创建充值订单失败");
  }
}
