import { NextRequest, NextResponse } from "next/server";
import { PaymentServiceFactory } from "@/lib/payment";
import { processPaymentCallback } from "@/lib/payment/callback";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { outTradeNo, success } = body;

    if (!outTradeNo) {
      return NextResponse.json(
        { success: false, error: "缺少订单号" },
        { status: 400 }
      );
    }

    console.debug("收到模拟支付回调:", body);

    const result = await PaymentServiceFactory.handleCallback("mock", {
      outTradeNo,
      success: success !== false,
    });

    if (result.status === "paid" && result.paidAt) {
      const callbackResult = await processPaymentCallback(
        result.outTradeNo,
        "paid",
        {
          transactionId: result.transactionId,
          paidAt: result.paidAt,
        }
      );

      if (!callbackResult.success) {
        console.error("处理支付回调失败:", callbackResult.error);
      }
    }

    return NextResponse.json({ success: true, data: result });
  } catch (error) {
    console.error("处理模拟支付回调失败:", error);
    return NextResponse.json(
      { success: false, error: "处理失败" },
      { status: 500 }
    );
  }
}
