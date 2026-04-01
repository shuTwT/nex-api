"use server";

import { revalidatePath } from "next/cache";
import { PaymentServiceFactory, getAvailablePaymentMethods, type PaymentMethod } from "@/lib/payment";
import { requireAuth } from "@/lib/auth/util";
import prisma from "@/lib/prisma";
import { authOptions } from "@/lib/auth/config";

export async function getPaymentMethods() {
  try {
    const methods = await getAvailablePaymentMethods();
    return { success: true, data: methods };
  } catch (error) {
    console.error("Error fetching payment methods:", error);
    return { success: false, error: "获取支付方式失败" };
  }
}

export async function createPayment(planId: string, method: PaymentMethod) {
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

    const notifyUrl = `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/business/subscription`;

    const result = await PaymentServiceFactory.createPayment({
      userId: sessionUser.id,
      amount: plan.price,
      method,
      notifyUrl,
      metadata: {
        type: "subscription",
        planId: plan.id,
      },
    });

    if (!result.success) {
      return { success: false, error: result.error || "创建支付失败" };
    }

    revalidatePath("/payment");
    return { success: true, data: result };
  } catch (error) {
    console.error("Error creating payment:", error);
    return { success: false, error: "创建支付失败" };
  }
}

export async function getPaymentInfo(outTradeNo: string) {
  try {
    await requireAuth(authOptions);

    const payment = await prisma.payment.findUnique({
      where: { outTradeNo },
      include: {
        user: true,
      },
    });

    if (!payment) {
      return { success: false, error: "支付记录不存在" };
    }

    return { success: true, data: payment };
  } catch (error) {
    console.error("Error fetching payment info:", error);
    return { success: false, error: "获取支付信息失败" };
  }
}

export async function queryPaymentStatus(outTradeNo: string) {
  try {
    await requireAuth(authOptions);

    const payment = await prisma.payment.findUnique({
      where: { outTradeNo },
    });

    if (!payment) {
      return { success: false, error: "支付记录不存在" };
    }

    const result = await PaymentServiceFactory.queryPayment(
      payment.method as PaymentMethod,
      outTradeNo
    );

    if (result && result.status !== payment.status) {
      await prisma.payment.update({
        where: { outTradeNo },
        data: {
          status: result.status,
          transactionId: result.transactionId,
          paidAt: result.paidAt,
        },
      });
    }

    return { success: true, data: result || payment };
  } catch (error) {
    console.error("Error querying payment status:", error);
    return { success: false, error: "查询支付状态失败" };
  }
}

export async function cancelPayment(outTradeNo: string) {
  try {
    await requireAuth(authOptions);

    const payment = await prisma.payment.findUnique({
      where: { outTradeNo },
    });

    if (!payment) {
      return { success: false, error: "支付记录不存在" };
    }

    if (payment.status !== "pending") {
      return { success: false, error: "只能取消待支付的订单" };
    }

    const result = await PaymentServiceFactory.closePayment(
      payment.method as PaymentMethod,
      outTradeNo
    );

    if (!result) {
      return { success: false, error: "取消支付失败" };
    }

    revalidatePath("/payment");
    return { success: true, message: "支付已取消" };
  } catch (error) {
    console.error("Error cancelling payment:", error);
    return { success: false, error: "取消支付失败" };
  }
}

export async function getUserPayments() {
  try {
    const sessionUser = await requireAuth(authOptions);

    const payments = await prisma.payment.findMany({
      where: { userId: sessionUser.id },
      orderBy: { createdAt: "desc" },
    });

    return { success: true, data: payments };
  } catch (error) {
    console.error("Error fetching user payments:", error);
    return { success: false, error: "获取支付记录失败" };
  }
}
