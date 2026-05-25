import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  try {
    const [totalServices, freeServices, totalCallCount] = await Promise.all([
      prisma.mcpService.count({ where: { isActive: true } }),
      prisma.mcpService.count({ where: { isActive: true, pricing: 0 } }),
      prisma.mcpService.aggregate({
        where: { isActive: true },
        _sum: { callCount: true },
      }),
    ]);

    return apiSuccess({
      totalServices,
      freeServices,
      paidServices: totalServices - freeServices,
      totalCallCount: totalCallCount._sum.callCount || 0,
    });
  } catch (error) {
    console.error("Error fetching MCP marketplace stats:", error);
    return apiError("获取 MCP 市场统计失败");
  }
}
