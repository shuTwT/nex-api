import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { PaymentServiceFactory, type PaymentMethod } from "@/lib/payment";

export async function POST(
  _request: NextRequest,
  { params }: { params: Promise<{ outTradeNo: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { outTradeNo } = await params;
    const payment = await prisma.payment.findUnique({ where: { outTradeNo } });
    if (!payment) return apiError("支付记录不存在", 404);
    if (payment.status !== "pending") return apiError("只能取消待支付的订单", 400);

    const result = await PaymentServiceFactory.closePayment(payment.method as PaymentMethod, outTradeNo);
    if (!result) return apiError("取消支付失败");

    revalidatePath("/payment");
    return apiSuccess({ message: "支付已取消" });
  } catch (error) {
    console.error("Error cancelling payment:", error);
    return apiError("取消支付失败");
  }
}
