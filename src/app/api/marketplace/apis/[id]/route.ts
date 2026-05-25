import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";
import redis from "@/lib/redis";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;
    const api = await prisma.api.findUnique({
      where: { id },
      include: { category: { select: { name: true } } },
    });

    if (!api) return apiError("API 不存在", 404);

    const todayCallCount = (async () => {
      try {
        const now = new Date();
        const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate());
        const timestamps: number[] = [];
        for (let i = 0; i < 24; i++) {
          const hour = new Date(startOfDay);
          hour.setHours(hour.getHours() + i);
          timestamps.push(Math.floor(hour.getTime() / 1000));
        }
        const keys = timestamps.map((ts) => `api:request:count:${api.alias}:hourly:${ts}`);
        const values = await redis.mget(...keys);
        return values.reduce((sum, value) => sum + parseInt(value || "0", 10), 0);
      } catch { return 0; }
    })();

    const userCount = (async () => {
      try {
        const keys = await redis.keys(`user:api:request:count:*:${api.alias}`);
        return keys.length;
      } catch { return 0; }
    })();

    const [today, users] = await Promise.all([todayCallCount, userCount]);

    return apiSuccess({
      id: api.id, name: api.name, description: api.description, alias: api.alias,
      endpoint: api.endpoint, method: api.method, pricing: api.pricing,
      category: api.category.name, isFree: api.pricing === 0, isActive: api.isActive,
      todayCallCount: today, userCount: users, totalCallCount: api.callCount,
      createdAt: api.createdAt, updatedAt: api.updatedAt,
    });
  } catch (error) {
    console.error("Error fetching API detail:", error);
    return apiError("获取 API 详情失败");
  }
}
