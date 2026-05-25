import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const [totalLogs, infoLogs, warningLogs, errorLogs, successLogs, failedLogs] = await Promise.all([
      prisma.auditLog.count(),
      prisma.auditLog.count({ where: { level: "info" } }),
      prisma.auditLog.count({ where: { level: "warning" } }),
      prisma.auditLog.count({ where: { level: "error" } }),
      prisma.auditLog.count({ where: { status: "success" } }),
      prisma.auditLog.count({ where: { status: "error" } }),
    ]);

    return apiSuccess({ totalLogs, infoLogs, warningLogs, errorLogs, successLogs, failedLogs });
  } catch (error) {
    console.error("Get audit log stats error:", error);
    return apiError("获取审计日志统计失败");
  }
}
