import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  try {
    const [totalApis, freeApis, totalCallCount] = await Promise.all([
      prisma.api.count({ where: { isActive: true } }),
      prisma.api.count({ where: { isActive: true, pricing: 0 } }),
      prisma.api.aggregate({ where: { isActive: true }, _sum: { callCount: true } }),
    ]);

    return apiSuccess({ totalApis, freeApis, paidApis: totalApis - freeApis, totalCallCount: totalCallCount._sum.callCount || 0 });
  } catch (error) {
    console.error("Error fetching marketplace stats:", error);
    return apiError("获取市场统计失败");
  }
}
