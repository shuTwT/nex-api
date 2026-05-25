import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const recentUsage = await prisma.apiUsage.findMany({
      where: { userId: user.id },
      include: { api: { select: { name: true, alias: true } } },
      orderBy: { createdAt: "desc" },
      take: 10,
    });

    const activities = recentUsage.map((usage) => ({
      id: usage.id, apiName: usage.api.name, apiAlias: usage.api.alias,
      credits: usage.credits, status: usage.status, createdAt: usage.createdAt,
    }));

    return apiSuccess(activities);
  } catch (error) {
    console.error("Error getting recent activity:", error);
    return apiError("获取最近活动失败");
  }
}
