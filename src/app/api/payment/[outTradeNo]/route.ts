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
    const payment = await prisma.payment.findUnique({ where: { outTradeNo }, include: { user: true } });
    if (!payment) return apiError("支付记录不存在", 404);
    return apiSuccess(payment);
  } catch (error) {
    console.error("Error fetching payment info:", error);
    return apiError("获取支付信息失败");
  }
}
