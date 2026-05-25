import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { searchParams } = new URL(request.url);
    const idsParam = searchParams.get("ids");
    const ids = idsParam ? idsParam.split(",") : undefined;

    const where = ids && ids.length > 0 ? { id: { in: ids } } : {};

    const codes = await prisma.redemptionCode.findMany({ where, orderBy: { createdAt: "desc" } });

    const typeLabels: Record<string, string> = { subscription: "订阅", quota: "额度" };
    const header = "\uFEFF兑换码,类型,订阅计划,额度,过期时间,状态,创建时间";
    const rows = codes.map((c) => {
      const typeLabel = typeLabels[c.type] || c.type;
      const planName = c.type === "subscription" ? c.planName || "-" : "-";
      const creditsVal = c.type === "quota" ? c.credits?.toString() || "-" : "-";
      const expires = c.expiresAt ? new Date(c.expiresAt).toLocaleString("zh-CN") : "永久";
      const status = c.isUsed ? "已使用" : "未使用";
      const createdAt = new Date(c.createdAt).toLocaleString("zh-CN");
      return [c.code, typeLabel, planName, creditsVal, expires, status, createdAt].map((v) => `"${v}"`).join(",");
    });

    return apiSuccess([header, ...rows].join("\n"));
  } catch (error) {
    console.error("Error exporting redemption codes:", error);
    return apiError("导出兑换码失败");
  }
}
