"use server";

import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/session";

export async function getPaymentSettings() {
  try {
    await requireAuth();

    const settings = await prisma.systemSetting.findMany({
      where: {
        category: "payment",
      },
    });

    const settingsMap: Record<string, string> = {};
    settings.forEach((s) => {
      settingsMap[s.key] = s.value;
    });

    return {
      success: true,
      data: {
        creditPrice: parseFloat(settingsMap.creditPrice || "1"),
        minRecharge: parseFloat(settingsMap.minRecharge || "10"),
        alipayEnabled: settingsMap.alipayEnabled === "true",
        wechatEnabled: settingsMap.wechatEnabled === "true",
      },
    };
  } catch (error) {
    console.error("获取支付设置失败:", error);
    return { success: false, error: "获取支付设置失败" };
  }
}
