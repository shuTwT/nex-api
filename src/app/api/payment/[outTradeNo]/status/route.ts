import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { PaymentServiceFactory, type PaymentMethod } from "@/lib/payment";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ outTradeNo: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { outTradeNo } = await params;
    const payment = await prisma.payment.findUnique({ where: { outTradeNo } });
    if (!payment) return apiError("支付记录不存在", 404);

    const result = await PaymentServiceFactory.queryPayment(payment.method as PaymentMethod, outTradeNo);

    if (result && result.status !== payment.status) {
      await prisma.payment.update({
        where: { outTradeNo },
        data: { status: result.status, transactionId: result.transactionId, paidAt: result.paidAt },
      });
    }

    return apiSuccess(result || payment);
  } catch (error) {
    console.error("Error querying payment status:", error);
    return apiError("查询支付状态失败");
  }
}
