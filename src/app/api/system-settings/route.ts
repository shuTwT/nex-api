import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";
import { clearPaymentConfigCache } from "@/lib/payment/config";
import { clearConfigCache } from "@/lib/config";

async function getCategoryFromKey(key: string): Promise<string> {
  const defaultSettings = {
    general: [
      { key: "siteName" }, { key: "siteDescription" }, { key: "siteLogo" }, { key: "contactEmail" },
    ],
    operation: {
      basic: [
        { key: "registrationEnabled" }, { key: "defaultCredits" }, { key: "inviteRewards" }, { key: "maintenanceMode" },
      ],
      announcement: [
        { key: "announcementEnabled" }, { key: "announcementContent" },
      ],
    },
    payment: {
      basic: [
        { key: "alipayEnabled" }, { key: "wechatEnabled" }, { key: "creditPrice" }, { key: "minRecharge" },
        { key: "mockPaymentEnabled" }, { key: "mockPaymentAutoSuccess" }, { key: "mockPaymentDelay" },
      ],
      alipay: [
        { key: "alipayAppId" }, { key: "alipayPrivateKey" }, { key: "alipayPublicKey" },
        { key: "alipayNotifyUrl" }, { key: "alipayReturnUrl" }, { key: "alipaySandbox" },
      ],
      wechat: [
        { key: "wechatPayAppId" }, { key: "wechatPayMchId" }, { key: "wechatPayApiKey" },
        { key: "wechatPayPrivateKey" }, { key: "wechatPayPublicKey" }, { key: "wechatPayPaymentPublicKey" },
        { key: "wechatPayPublicKeyId" }, { key: "wechatPayNotifyUrl" }, { key: "wechatPayDebug" },
      ],
    },
    oauth: {
      basic: [{ key: "oauthProviders" }],
    },
  };

  for (const [category, settings] of Object.entries(defaultSettings)) {
    if (category === "payment" || category === "operation" || category === "oauth") {
      for (const s of Object.values(settings as Record<string, Array<{ key: string }>>)) {
        if (s.some((setting) => setting.key === key)) return category;
      }
    } else if ((settings as Array<{ key: string }>).some((s) => s.key === key)) {
      return category;
    }
  }
  return "general";
}

export async function GET(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { searchParams } = new URL(request.url);
    const category = searchParams.get("category") || undefined;

    const where: Record<string, unknown> = {};
    if (category) where.category = category;

    const settings = await prisma.systemSetting.findMany({ where, orderBy: { key: "asc" } });
    return apiSuccess(settings);
  } catch (error) {
    console.error("Get system settings error:", error);
    return apiError("获取系统设置失败");
  }
}

export async function PUT(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { settings }: { settings: Array<{ key: string; value: string }> } = body;

    for (const setting of settings) {
      await prisma.systemSetting.upsert({
        where: { key: setting.key },
        update: { value: setting.value },
        create: { key: setting.key, value: setting.value, category: await getCategoryFromKey(setting.key) },
      });
    }

    await logAudit({ action: "更新系统设置", resource: "系统设置", details: `更新了 ${settings.length} 个系统设置`, level: "info", status: "success" });

    const hasPaymentSettings = (await Promise.all(settings.map(async s => await getCategoryFromKey(s.key)))).some(c => c === "payment");
    if (hasPaymentSettings) clearPaymentConfigCache();

    clearConfigCache();
    revalidatePath("/console/settings");
    return apiSuccess({ message: "设置已更新" });
  } catch (error) {
    console.error("Update system settings error:", error);
    await logAudit({ action: "更新系统设置", resource: "系统设置", details: `更新系统设置失败: ${error}`, level: "error", status: "error" });
    return apiError("更新系统设置失败");
  }
}
