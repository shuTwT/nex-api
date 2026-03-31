"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/session";
import { logAudit } from "@/lib/audit-log";

export async function getSystemSettings(category?: string) {
  try {
    await requireAdmin();

    const where: any = {};
    if (category) {
      where.category = category;
    }

    const settings = await prisma.systemSetting.findMany({
      where,
      orderBy: {
        key: "asc",
      },
    });

    return { success: true, data: settings };
  } catch (error) {
    console.error("Get system settings error:", error);
    return { success: false, error: "获取系统设置失败" };
  }
}

export async function updateSystemSettings(settings: Array<{ key: string; value: string }>) {
  try {
    await requireAdmin();

    for (const setting of settings) {
      await prisma.systemSetting.upsert({
        where: { key: setting.key },
        update: { value: setting.value },
        create: {
          key: setting.key,
          value: setting.value,
          category: getCategoryFromKey(setting.key),
        },
      });
    }

    await logAudit({
      action: "更新系统设置",
      resource: "系统设置",
      details: `更新了 ${settings.length} 个系统设置`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/settings");
    return { success: true };
  } catch (error) {
    console.error("Update system settings error:", error);
    await logAudit({
      action: "更新系统设置",
      resource: "系统设置",
      details: `更新系统设置失败: ${error}`,
      level: "error",
      status: "error",
    });
    return { success: false, error: "更新系统设置失败" };
  }
}

export async function getDefaultSettings() {
  return {
    general: [
      { key: "siteName", value: "API 网关", category: "general", description: "网站名称" },
      { key: "siteDescription", value: "一站式 API 服务平台", category: "general", description: "网站描述" },
      { key: "siteLogo", value: "", category: "general", description: "网站 Logo" },
      { key: "contactEmail", value: "support@example.com", category: "general", description: "联系邮箱" },
    ],
    operation: [
      { key: "registrationEnabled", value: "true", category: "operation", description: "是否允许用户注册" },
      { key: "defaultCredits", value: "1000", category: "operation", description: "新用户默认积分" },
      { key: "inviteRewards", value: "100", category: "operation", description: "邀请奖励积分" },
      { key: "maintenanceMode", value: "false", category: "operation", description: "是否开启维护模式" },
    ],
    payment: [
      { key: "alipayEnabled", value: "false", category: "payment", description: "是否开启支付宝支付" },
      { key: "wechatEnabled", value: "false", category: "payment", description: "是否开启微信支付" },
      { key: "creditPrice", value: "1", category: "payment", description: "每积分价格（元）" },
      { key: "minRecharge", value: "10", category: "payment", description: "最低充值金额（元）" },
    ],
  };
}

function getCategoryFromKey(key: string): string {
  const defaultSettings = getDefaultSettings();
  for (const [category, settings] of Object.entries(defaultSettings)) {
    if (settings.some((s: { key: string }) => s.key === key)) {
      return category;
    }
  }
  return "general";
}
