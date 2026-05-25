import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const [totalUsers, activeUsers, adminUsers, newUsersThisMonth] = await Promise.all([
      prisma.user.count(),
      prisma.user.count({ where: { apiUsage: { some: { createdAt: { gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000) } } } } }),
      prisma.user.count({ where: { role: "admin" } }),
      prisma.user.count({ where: { createdAt: { gte: new Date(new Date().getFullYear(), new Date().getMonth(), 1) } } }),
    ]);

    return apiSuccess({ totalUsers, activeUsers, adminUsers, newUsersThisMonth });
  } catch (error) {
    console.error("Error fetching user stats:", error);
    return apiError("获取用户统计失败");
  }
}
