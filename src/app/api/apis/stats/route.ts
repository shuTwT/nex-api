import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const [totalApis, activeApis, inactiveApis, totalCalls, categoriesCount] = await Promise.all([
      prisma.api.count(),
      prisma.api.count({ where: { isActive: true } }),
      prisma.api.count({ where: { isActive: false } }),
      prisma.api.aggregate({ _sum: { callCount: true } }),
      prisma.apiCategory.count(),
    ]);

    return apiSuccess({
      totalApis, activeApis, inactiveApis,
      totalCalls: totalCalls._sum.callCount || 0, categoriesCount,
    });
  } catch (error) {
    console.error("Get API stats error:", error);
    return apiError("获取 API 统计失败");
  }
}
