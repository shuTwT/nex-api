import { NextRequest, NextResponse } from "next/server";
import { PaymentServiceFactory } from "@/lib/payment";
import { processPaymentCallback } from "@/lib/payment/callback";

export async function POST(request: NextRequest) {
  try {
    const body = await request.text();
    const headers = Object.fromEntries(request.headers.entries());

    const result = await PaymentServiceFactory.handleCallback("wechat", {
      headers,
      body,
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
        console.error("处理微信支付回调失败:", callbackResult.error);
      }
    }

    return new NextResponse(
      JSON.stringify({
        code: "SUCCESS",
        message: "成功",
      }),
      {
        status: 200,
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
  } catch (error) {
    console.error("处理微信支付回调失败:", error);
    return new NextResponse(
      JSON.stringify({
        code: "FAIL",
        message: "失败",
      }),
      {
        status: 500,
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
  }
}
