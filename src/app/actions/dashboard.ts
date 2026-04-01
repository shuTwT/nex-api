"use server";

import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/auth/util";
import redis from "@/lib/redis";
import { getHourlyUsageTrend } from "@/lib/request-stats";
import { authOptions } from "@/lib/auth/config";

export async function getDashboardStats() {
  try {
    const user = await requireAuth(authOptions);

    const now = new Date();
    const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);

    const monthlyUsage = await prisma.apiUsage.aggregate({
      where: {
        userId: user.id,
        createdAt: {
          gte: firstDayOfMonth,
        },
      },
      _sum: {
        credits: true,
      },
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
        if (calls > 0) {
          activeApis.add(alias);
        }
      });
    }

    const userInfo = await prisma.user.findUnique({
      where: { id: user.id },
      select: { credits: true },
    });

    return {
      success: true,
      data: {
        monthlyCreditsUsed: monthlyUsage._sum.credits || 0,
        apiCalls: totalApiCalls,
        accountBalance: userInfo?.credits || 0,
        activeApis: activeApis.size,
      },
    };
  } catch (error) {
    console.error("Error getting dashboard stats:", error);
    return { success: false, error: "获取概览数据失败" };
  }
}

export async function getRecentActivity() {
  try {
    const user = await requireAuth(authOptions);

    const recentUsage = await prisma.apiUsage.findMany({
      where: {
        userId: user.id,
      },
      include: {
        api: {
          select: {
            name: true,
            alias: true,
          },
        },
      },
      orderBy: {
        createdAt: "desc",
      },
      take: 10,
    });

    const activities = recentUsage.map((usage) => ({
      id: usage.id,
      apiName: usage.api.name,
      apiAlias: usage.api.alias,
      credits: usage.credits,
      status: usage.status,
      createdAt: usage.createdAt,
    }));

    return { success: true, data: activities };
  } catch (error) {
    console.error("Error getting recent activity:", error);
    return { success: false, error: "获取最近活动失败" };
  }
}

export async function getTopApis() {
  try {
    const user = await requireAuth(authOptions);

    const userApiKeys = await redis.keys(`user:api:request:count:${user.id}:*`);
    
    if (userApiKeys.length === 0) {
      return { success: true, data: [] };
    }

    const values = await redis.mget(...userApiKeys);
    const apiCalls: { alias: string; calls: number }[] = [];

    userApiKeys.forEach((key, index) => {
      const alias = key.replace(`user:api:request:count:${user.id}:`, "");
      const calls = parseInt(values[index] || "0", 10);
      apiCalls.push({ alias, calls });
    });

    apiCalls.sort((a, b) => b.calls - a.calls);
    const topApiCalls = apiCalls.slice(0, 5);
    const totalCalls = topApiCalls.reduce((sum, api) => sum + api.calls, 0);

    const apiDetails = await prisma.api.findMany({
      where: {
        alias: {
          in: topApiCalls.map((api) => api.alias),
        },
      },
      select: {
        name: true,
        alias: true,
      },
    });

    const result = topApiCalls.map((api) => {
      const details = apiDetails.find((d) => d.alias === api.alias);
      return {
        name: details?.name || api.alias,
        calls: api.calls,
        percentage: totalCalls > 0 ? Math.round((api.calls / totalCalls) * 100) : 0,
      };
    });

    return { success: true, data: result };
  } catch (error) {
    console.error("Error getting top APIs:", error);
    return { success: false, error: "获取热门接口失败" };
  }
}

export async function getUsageTrend() {
  try {
    const user = await requireAuth(authOptions);
    const isAdmin = user.role === "admin";

    const now = new Date();
    const labels: string[] = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date(now);
      d.setHours(d.getHours() - i);
      labels.push(`${d.getHours()}:00`);
    }

    const userTrend = await getHourlyUsageTrend(user.id, 7);

    if (isAdmin) {
      const globalTrend = await getHourlyUsageTrend(undefined, 7);
      return {
        success: true,
        data: {
          labels,
          datasets: [
            {
              label: "全局用量",
              data: globalTrend,
              borderColor: "rgb(59, 130, 246)",
              backgroundColor: "rgba(59, 130, 246, 0.1)",
            },
            {
              label: "我的用量",
              data: userTrend,
              borderColor: "rgb(16, 185, 129)",
              backgroundColor: "rgba(16, 185, 129, 0.1)",
            },
          ],
        },
      };
    }

    return {
      success: true,
      data: {
        labels,
        datasets: [
          {
            label: "我的用量",
            data: userTrend,
            borderColor: "rgb(59, 130, 246)",
            backgroundColor: "rgba(59, 130, 246, 0.1)",
          },
        ],
      },
    };
  } catch (error) {
    console.error("Error getting usage trend:", error);
    return { success: false, error: "获取用量趋势失败" };
  }
}
