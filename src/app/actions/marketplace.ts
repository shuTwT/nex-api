"use server";

import prisma from "@/lib/prisma";
import redis from "@/lib/redis";

export async function getMarketplaceApis() {
  try {
    const apis = await prisma.api.findMany({
      where: {
        isActive: true,
      },
      include: {
        category: {
          select: {
            name: true,
          },
        },
      },
      orderBy: {
        callCount: "desc",
      },
      take: 20,
    });

    const apiStats = await Promise.all(
      apis.map(async (api) => {
        const todayCallCount = await getTodayApiCallCount(api.alias);
        const userCount = await getApiUserCount(api.alias);

        return {
          id: api.id,
          name: api.name,
          description: api.description,
          alias: api.alias,
          endpoint: api.endpoint,
          method: api.method,
          pricing: api.pricing,
          category: api.category.name,
          isFree: api.pricing === 0,
          todayCallCount,
          userCount,
          totalCallCount: api.callCount,
        };
      })
    );

    return { success: true, data: apiStats };
  } catch (error) {
    console.error("Error fetching marketplace APIs:", error);
    return { success: false, error: "获取 API 列表失败" };
  }
}

export async function getApiDetail(id: string) {
  try {
    const api = await prisma.api.findUnique({
      where: { id },
      include: {
        category: {
          select: {
            name: true,
          },
        },
      },
    });

    if (!api) {
      return { success: false, error: "API 不存在" };
    }

    const todayCallCount = await getTodayApiCallCount(api.alias);
    const userCount = await getApiUserCount(api.alias);

    return {
      success: true,
      data: {
        id: api.id,
        name: api.name,
        description: api.description,
        alias: api.alias,
        endpoint: api.endpoint,
        method: api.method,
        pricing: api.pricing,
        category: api.category.name,
        isFree: api.pricing === 0,
        isActive: api.isActive,
        todayCallCount,
        userCount,
        totalCallCount: api.callCount,
        createdAt: api.createdAt,
        updatedAt: api.updatedAt,
      },
    };
  } catch (error) {
    console.error("Error fetching API detail:", error);
    return { success: false, error: "获取 API 详情失败" };
  }
}

async function getTodayApiCallCount(apiAlias: string): Promise<number> {
  try {
    const now = new Date();
    const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const timestamps: number[] = [];

    for (let i = 0; i < 24; i++) {
      const hour = new Date(startOfDay);
      hour.setHours(hour.getHours() + i);
      timestamps.push(Math.floor(hour.getTime() / 1000));
    }

    const keys = timestamps.map((ts) => `api:request:count:${apiAlias}:hourly:${ts}`);
    const values = await redis.mget(...keys);

    return values.reduce((sum, value) => {
      return sum + parseInt(value || "0", 10);
    }, 0);
  } catch (error) {
    console.error("Error getting today's API call count:", error);
    return 0;
  }
}

async function getApiUserCount(apiAlias: string): Promise<number> {
  try {
    const keys = await redis.keys(`user:api:request:count:*:${apiAlias}`);
    return keys.length;
  } catch (error) {
    console.error("Error getting API user count:", error);
    return 0;
  }
}

export async function getMarketplaceStats() {
  try {
    const [totalApis, freeApis, totalCallCount] = await Promise.all([
      prisma.api.count({ where: { isActive: true } }),
      prisma.api.count({ where: { isActive: true, pricing: 0 } }),
      prisma.api.aggregate({
        where: { isActive: true },
        _sum: { callCount: true },
      }),
    ]);

    return {
      success: true,
      data: {
        totalApis,
        freeApis,
        paidApis: totalApis - freeApis,
        totalCallCount: totalCallCount._sum.callCount || 0,
      },
    };
  } catch (error) {
    console.error("Error fetching marketplace stats:", error);
    return { success: false, error: "获取市场统计失败" };
  }
}
