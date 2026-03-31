import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import type { BusinessCallbackPayload } from "@/lib/payment/callback";

export async function POST(request: NextRequest) {
  try {
    const payload: BusinessCallbackPayload = await request.json();

    console.log("收到充值业务回调:", payload);

    if (payload.status !== "paid") {
      return NextResponse.json({ 
        success: true, 
        message: "非支付成功状态，跳过处理" 
      });
    }

    if (!payload.metadata) {
      return NextResponse.json(
        { success: false, error: "缺少充值元数据" },
        { status: 400 }
      );
    }

    const metadata = typeof payload.metadata === 'string' 
      ? JSON.parse(payload.metadata) 
      : payload.metadata;

    if (metadata.type !== "recharge") {
      return NextResponse.json(
        { success: false, error: "非充值类型支付" },
        { status: 400 }
      );
    }

    const credits = metadata.credits;
    if (!credits || typeof credits !== "number") {
      return NextResponse.json(
        { success: false, error: "无效的积分数量" },
        { status: 400 }
      );
    }

    await prisma.$transaction(async (tx) => {
      await tx.user.update({
        where: { id: payload.userId },
        data: {
          credits: {
            increment: credits,
          },
        },
      });

      console.log(`用户 ${payload.userId} 充值 ${credits} 积分成功`);
    });

    return NextResponse.json({ 
      success: true, 
      message: "充值成功" 
    });
  } catch (error) {
    console.error("处理充值业务回调失败:", error);
    return NextResponse.json(
      { 
        success: false, 
        error: error instanceof Error ? error.message : "处理失败" 
      },
      { status: 500 }
    );
  }
}
