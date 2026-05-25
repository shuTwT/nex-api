import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const last7Days = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    const last30Days = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);

    const [userRecord, totalUsage, todayUsage, last7DaysUsage, last30DaysUsage] = await Promise.all([
      prisma.user.findUnique({ where: { id: user.id }, select: { credits: true } }),
      prisma.apiUsage.count(),
      prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: todayStart } } }),
      prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: last7Days } } }),
      prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: last30Days } } }),
    ]);

    const todayHourlyUsage = await Promise.all(
      Array.from({ length: 24 }, (_, hour) => {
        const hourStart = new Date(todayStart.getTime() + hour * 60 * 60 * 1000);
        const hourEnd = new Date(hourStart.getTime() + 60 * 60 * 1000);
        return prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: hourStart, lt: hourEnd } } });
      })
    );

    const last7DaysDailyUsage = await Promise.all(
      Array.from({ length: 7 }, (_, day) => {
        const dayStart = new Date(last7Days.getTime() + day * 24 * 60 * 60 * 1000);
        const dayEnd = new Date(dayStart.getTime() + 24 * 60 * 60 * 1000);
        return prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: dayStart, lt: dayEnd } } });
      })
    );

    const last30DaysDailyUsage = await Promise.all(
      Array.from({ length: 30 }, (_, day) => {
        const dayStart = new Date(last30Days.getTime() + day * 24 * 60 * 60 * 1000);
        const dayEnd = new Date(dayStart.getTime() + 24 * 60 * 60 * 1000);
        return prisma.apiUsage.count({ where: { userId: user.id, createdAt: { gte: dayStart, lt: dayEnd } } });
      })
    );

    return apiSuccess({
      freeCredits: userRecord?.credits || 0,
      totalUsage, todayUsage, last7DaysUsage, last30DaysUsage,
      todayHourlyUsage, last7DaysDailyUsage, last30DaysDailyUsage,
    });
  } catch (error) {
    console.error("Get usage stats error:", error);
    return apiError("获取用量统计失败");
  }
}
