import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import redis from "@/lib/redis";
import { getHourlyUsageTrend } from "@/lib/request-stats";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const now = new Date();
    const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);

    const monthlyUsage = await prisma.apiUsage.aggregate({
      where: { userId: user.id, createdAt: { gte: firstDayOfMonth } },
      _sum: { credits: true },
    });

    const userApiKeys = await redis.keys(`user:api:request:count:${user.id}:*`);
    let totalApiCalls = 0;
    const activeApis = new Set<string>();

    if (userApiKeys.length > 0) {
      const values = await redis.mget(...userApiKeys);
      userApiKeys.forEach((key, index) => {
        const alias = key.replace(`user:api:request:count:${user.id}:`, "");
        const calls = parseInt(values[index] || "0", 10);
        totalApiCalls += calls;
        if (calls > 0) activeApis.add(alias);
      });
    }

    const userInfo = await prisma.user.findUnique({ where: { id: user.id }, select: { credits: true } });

    return apiSuccess({
      monthlyCreditsUsed: monthlyUsage._sum.credits || 0,
      apiCalls: totalApiCalls,
      accountBalance: userInfo?.credits || 0,
      activeApis: activeApis.size,
    });
  } catch (error) {
    console.error("Error getting dashboard stats:", error);
    return apiError("获取概览数据失败");
  }
}
