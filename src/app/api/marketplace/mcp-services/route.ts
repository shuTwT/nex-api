import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";
import redis from "@/lib/redis";

export async function GET(_request: NextRequest) {
  try {
    const services = await prisma.mcpService.findMany({
      where: { isActive: true },
      orderBy: { callCount: "desc" },
      take: 20,
    });

    const serviceStats = await Promise.all(
      services.map(async (svc) => {
        const todayCallCount = (async () => {
          try {
            const now = new Date();
            const startOfDay = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate()
            );
            const timestamps: number[] = [];
            for (let i = 0; i < 24; i++) {
              const hour = new Date(startOfDay);
              hour.setHours(hour.getHours() + i);
              timestamps.push(Math.floor(hour.getTime() / 1000));
            }
            const keys = timestamps.map(
              (ts) => `mcp:request:count:mcp:${svc.identifier}:hourly:${ts}`
            );
            const values = await redis.mget(...keys);
            return values.reduce(
              (sum, value) => sum + parseInt(value || "0", 10),
              0
            );
          } catch {
            return 0;
          }
        })();

        const userCount = (async () => {
          try {
            const keys = await redis.keys(
              `user:mcp:request:count:*:mcp:${svc.identifier}`
            );
            return keys.length;
          } catch {
            return 0;
          }
        })();

        const [today, users] = await Promise.all([todayCallCount, userCount]);

        return {
          id: svc.id,
          name: svc.name,
          identifier: svc.identifier,
          type: svc.type,
          pricing: svc.pricing,
          isFree: svc.pricing === 0,
          todayCallCount: today,
          userCount: users,
          totalCallCount: svc.callCount,
        };
      })
    );

    return apiSuccess(serviceStats);
  } catch (error) {
    console.error("Error fetching marketplace MCP services:", error);
    return apiError("获取 MCP 服务列表失败");
  }
}
