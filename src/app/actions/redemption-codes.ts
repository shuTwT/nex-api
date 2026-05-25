"use server";

import { revalidatePath } from "next/cache";
import crypto from "crypto";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/auth/util";
import { authOptions } from "@/lib/auth/config";

function generateCode(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let code = "";
  const bytes = crypto.randomBytes(12);
  for (let i = 0; i < 12; i++) {
    code += chars[bytes[i] % chars.length];
  }
  return code;
}

export async function getRedemptionCodes(params: {
  page?: number;
  limit?: number;
}) {
  try {
    const { page = 1, limit = 20 } = params;
    const skip = (page - 1) * limit;

    const [codes, total] = await Promise.all([
      prisma.redemptionCode.findMany({
        orderBy: { createdAt: "desc" },
        skip,
        take: limit,
      }),
      prisma.redemptionCode.count(),
    ]);

    return {
      success: true,
      data: codes,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Error fetching redemption codes:", error);
    return { success: false, error: "获取兑换码失败" };
  }
}

export async function createRedemptionCodes(formData: FormData) {
  try {
    const sessionUser = await requireAdmin(authOptions);

    const type = formData.get("type") as string;
    const count = parseInt(formData.get("count") as string) || 1;
    const planId = formData.get("planId") as string | null;
    const credits = formData.get("credits")
      ? parseInt(formData.get("credits") as string)
      : null;
    const expiresAtStr = formData.get("expiresAt") as string | null;

    if (!type || (type !== "subscription" && type !== "quota")) {
      return { success: false, error: "请选择有效的兑换码类型" };
    }

    if (type === "subscription" && !planId) {
      return { success: false, error: "请选择订阅计划" };
    }

    if (type === "quota" && (!credits || credits <= 0)) {
      return { success: false, error: "请输入有效额度" };
    }

    if (count < 1 || count > 1000) {
      return { success: false, error: "生成数量范围为 1-1000" };
    }

    const batchId = crypto.randomUUID();
    const codes: string[] = [];

    for (let i = 0; i < count; i++) {
      let code: string;
      let attempts = 0;
      do {
        code = generateCode();
        attempts++;
        if (attempts > 10) {
          return { success: false, error: "生成兑换码失败，请重试" };
        }
      } while (codes.includes(code));

      codes.push(code);
    }

    const existingCodes = await prisma.redemptionCode.findMany({
      where: { code: { in: codes } },
      select: { code: true },
    });

    if (existingCodes.length > 0) {
      return { success: false, error: "生成的兑换码与已有兑换码冲突，请重试" };
    }

    let planName: string | null = null;
    if (type === "subscription" && planId) {
      const plan = await prisma.subscriptionPlan.findUnique({
        where: { id: planId },
        select: { title: true },
      });
      planName = plan?.title ?? null;
    }

    const expiresAt = expiresAtStr ? new Date(expiresAtStr) : null;

    await prisma.redemptionCode.createMany({
      data: codes.map((code) => ({
        code,
        type,
        planId: type === "subscription" ? planId : null,
        planName,
        credits: type === "quota" ? credits : null,
        expiresAt,
        createdBy: sessionUser.id,
        batchId,
      })),
    });

    revalidatePath("/console/redemption-codes");
    return { success: true, data: { count, batchId } };
  } catch (error) {
    console.error("Error creating redemption codes:", error);
    return { success: false, error: "创建兑换码失败" };
  }
}

export async function deleteRedemptionCode(id: string) {
  try {
    await requireAdmin(authOptions);

    const code = await prisma.redemptionCode.findUnique({
      where: { id },
    });

    if (!code) {
      return { success: false, error: "兑换码未找到" };
    }

    if (code.isUsed) {
      return { success: false, error: "该兑换码已被使用，无法删除" };
    }

    await prisma.redemptionCode.delete({
      where: { id },
    });

    revalidatePath("/console/redemption-codes");
    return { success: true, message: "兑换码已删除" };
  } catch (error) {
    console.error("Error deleting redemption code:", error);
    return { success: false, error: "删除兑换码失败" };
  }
}

export async function deleteRedemptionCodeBatch(batchId: string) {
  try {
    await requireAdmin(authOptions);

    const usedCount = await prisma.redemptionCode.count({
      where: { batchId, isUsed: true },
    });

    if (usedCount > 0) {
      return {
        success: false,
        error: `该批次有 ${usedCount} 个兑换码已被使用，无法批量删除`,
      };
    }

    const result = await prisma.redemptionCode.deleteMany({
      where: { batchId, isUsed: false },
    });

    revalidatePath("/console/redemption-codes");
    return {
      success: true,
      message: `已删除 ${result.count} 个兑换码`,
    };
  } catch (error) {
    console.error("Error batch deleting redemption codes:", error);
    return { success: false, error: "批量删除兑换码失败" };
  }
}

export async function deleteRedemptionCodes(ids: string[]) {
  try {
    await requireAdmin(authOptions);

    if (!ids || ids.length === 0) {
      return { success: false, error: "请选择要删除的兑换码" };
    }

    const usedCount = await prisma.redemptionCode.count({
      where: { id: { in: ids }, isUsed: true },
    });

    if (usedCount > 0) {
      return {
        success: false,
        error: `所选中有 ${usedCount} 个兑换码已被使用，无法删除`,
      };
    }

    const result = await prisma.redemptionCode.deleteMany({
      where: { id: { in: ids }, isUsed: false },
    });

    revalidatePath("/console/redemption-codes");
    return { success: true, message: `已删除 ${result.count} 个兑换码` };
  } catch (error) {
    console.error("Error batch deleting redemption codes:", error);
    return { success: false, error: "批量删除兑换码失败" };
  }
}

export async function exportRedemptionCodes(ids?: string[]) {
  try {
    await requireAdmin(authOptions);

    const where = ids && ids.length > 0 ? { id: { in: ids } } : {};

    const codes = await prisma.redemptionCode.findMany({
      where,
      orderBy: { createdAt: "desc" },
    });

    const typeLabels: Record<string, string> = {
      subscription: "订阅",
      quota: "额度",
    };

    const header =
      "\uFEFF兑换码,类型,订阅计划,额度,过期时间,状态,创建时间";
    const rows = codes.map((c) => {
      const typeLabel = typeLabels[c.type] || c.type;
      const planName = c.type === "subscription" ? c.planName || "-" : "-";
      const creditsVal = c.type === "quota" ? c.credits?.toString() || "-" : "-";
      const expires = c.expiresAt
        ? new Date(c.expiresAt).toLocaleString("zh-CN")
        : "永久";
      const status = c.isUsed ? "已使用" : "未使用";
      const createdAt = new Date(c.createdAt).toLocaleString("zh-CN");

      return [c.code, typeLabel, planName, creditsVal, expires, status, createdAt]
        .map((v) => `"${v}"`)
        .join(",");
    });

    const csv = [header, ...rows].join("\n");

    return { success: true, data: csv };
  } catch (error) {
    console.error("Error exporting redemption codes:", error);
    return { success: false, error: "导出兑换码失败" };
  }
}

export async function getSubscriptionPlansForSelect() {
  try {
    const plans = await prisma.subscriptionPlan.findMany({
      where: { isActive: true },
      orderBy: { sortOrder: "asc" },
      select: { id: true, title: true },
    });

    return { success: true, data: plans };
  } catch (error) {
    console.error("Error fetching subscription plans:", error);
    return { success: false, error: "获取订阅计划失败" };
  }
}
