import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { searchParams } = new URL(request.url);
    const level = searchParams.get("level") || undefined;
    const status = searchParams.get("status") || undefined;
    const startDate = searchParams.get("startDate") || undefined;
    const endDate = searchParams.get("endDate") || undefined;

    const where: Record<string, unknown> = {};
    if (level && level !== "all") where.level = level;
    if (status && status !== "all") where.status = status;
    if (startDate || endDate) {
      where.createdAt = {} as Record<string, Date>;
      if (startDate) (where.createdAt as Record<string, Date>).gte = new Date(startDate);
      if (endDate) (where.createdAt as Record<string, Date>).lte = new Date(endDate);
    }

    const logs = await prisma.auditLog.findMany({
      where,
      orderBy: { createdAt: "desc" },
      include: { user: { select: { id: true, name: true, email: true } } },
    });

    const csvHeader = "时间,用户,操作,资源,详情,IP地址,级别,状态\n";
    const csvRows = logs.map(log => {
      const userName = log.user ? log.user.email || log.user.name || "系统" : "系统";
      return [log.createdAt.toISOString(), userName, log.action, log.resource, log.details || "", log.ipAddress || "", log.level, log.status].join(",");
    }).join("\n");

    return apiSuccess(csvHeader + csvRows);
  } catch (error) {
    console.error("Export audit logs error:", error);
    return apiError("导出审计日志失败");
  }
}
