"use server";

import { revalidatePath } from "next/cache";
import { PaymentServiceFactory, type PaymentMethod } from "@/lib/payment";
import { requireAuth } from "@/lib/auth/util";
import prisma from "@/lib/prisma";
import { authOptions } from "@/lib/auth/config";

export async function createRechargePayment(params: {
  amount: number;
  credits: number;
  method: PaymentMethod;
}) {
  try {
    const sessionUser = await requireAuth(authOptions);

    const settings = await prisma.systemSetting.findMany({
      where: {
        category: "payment",
      },
    });

    const settingsMap: Record<string, string> = {};
    settings.forEach((s) => {
      settingsMap[s.key] = s.value;
    });

    const minRecharge = parseFloat(settingsMap.minRecharge || "10");
    const creditPrice = parseFloat(settingsMap.creditPrice || "1");

    if (params.amount < minRecharge) {
      return { success: false, error: `最低充值金额为 ¥${minRecharge}` };
    }

    const expectedCredits = Math.floor(params.amount / creditPrice);
    if (params.credits !== expectedCredits) {
      return { success: false, error: "积分计算错误" };
    }

    const notifyUrl = `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/business/recharge`;

    const result = await PaymentServiceFactory.createPayment({
      userId: sessionUser.id,
      planId: undefined,
      amount: params.amount,
      method: params.method,
      notifyUrl,
      metadata: {
        type: "recharge",
        credits: params.credits,
        creditPrice,
      },
    });

    if (!result.success) {
      return { success: false, error: result.error || "创建支付失败" };
    }

    revalidatePath("/payment");
    return { success: true, data: result };
  } catch (error) {
    console.error("Error creating recharge payment:", error);
    return { success: false, error: "创建充值订单失败" };
  }
}
