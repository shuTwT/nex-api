import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { isMockPaymentEnabled } from "@/lib/payment/config";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const settings = await prisma.systemSetting.findMany({ where: { category: "payment" } });

    const settingsMap: Record<string, string> = {};
    settings.forEach((s) => { settingsMap[s.key] = s.value; });

    return apiSuccess({
      creditPrice: parseFloat(settingsMap.creditPrice || "1"),
      minRecharge: parseFloat(settingsMap.minRecharge || "10"),
      alipayEnabled: settingsMap.alipayEnabled === "true",
      wechatEnabled: settingsMap.wechatEnabled === "true",
      mockEnabled: await isMockPaymentEnabled(),
    });
  } catch (error) {
    console.error("获取支付设置失败:", error);
    return apiError("获取支付设置失败");
  }
}
