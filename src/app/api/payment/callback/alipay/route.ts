import { NextRequest, NextResponse } from "next/server";
import { PaymentServiceFactory } from "@/lib/payment";
import { processPaymentCallback } from "@/lib/payment/callback";

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData();
    const data: Record<string, string> = {};
    
    formData.forEach((value, key) => {
      data[key] = value.toString();
    });

    const result = await PaymentServiceFactory.handleCallback("alipay", data);

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
        console.error("处理支付宝支付回调失败:", callbackResult.error);
      }
    }

    return new NextResponse("success", {
      status: 200,
      headers: {
        "Content-Type": "text/plain",
      },
    });
  } catch (error) {
    console.error("处理支付宝支付回调失败:", error);
    return new NextResponse("fail", {
      status: 500,
      headers: {
        "Content-Type": "text/plain",
      },
    });
  }
}
