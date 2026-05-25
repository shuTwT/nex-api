import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import redis from "@/lib/redis";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const userApiKeys = await redis.keys(`user:api:request:count:${user.id}:*`);
    if (userApiKeys.length === 0) return apiSuccess([]);

    const values = await redis.mget(...userApiKeys);
    const apiCalls: { alias: string; calls: number }[] = [];

    userApiKeys.forEach((key, index) => {
      const alias = key.replace(`user:api:request:count:${user.id}:`, "");
      apiCalls.push({ alias, calls: parseInt(values[index] || "0", 10) });
    });

    apiCalls.sort((a, b) => b.calls - a.calls);
    const topApiCalls = apiCalls.slice(0, 5);
    const totalCalls = topApiCalls.reduce((sum, api) => sum + api.calls, 0);

    const apiDetails = await prisma.api.findMany({
      where: { alias: { in: topApiCalls.map((api) => api.alias) } },
      select: { name: true, alias: true },
    });

    const result = topApiCalls.map((api) => {
      const details = apiDetails.find((d) => d.alias === api.alias);
      return { name: details?.name || api.alias, calls: api.calls, percentage: totalCalls > 0 ? Math.round((api.calls / totalCalls) * 100) : 0 };
    });

    return apiSuccess(result);
  } catch (error) {
    console.error("Error getting top APIs:", error);
    return apiError("获取热门接口失败");
  }
}
