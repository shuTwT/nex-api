import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const [totalServices, activeServices, inactiveServices, totalCalls] =
      await Promise.all([
        prisma.mcpService.count(),
        prisma.mcpService.count({ where: { isActive: true } }),
        prisma.mcpService.count({ where: { isActive: false } }),
        prisma.mcpService.aggregate({ _sum: { callCount: true } }),
      ]);

    return apiSuccess({
      totalServices,
      activeServices,
      inactiveServices,
      totalCalls: totalCalls._sum.callCount || 0,
    });
  } catch (error) {
    console.error("Get MCP stats error:", error);
    return apiError("获取 MCP 统计失败");
  }
}
