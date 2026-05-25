import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const now = new Date();
    const [totalTokens, activeTokens, inactiveTokens, expiredTokens] = await Promise.all([
      prisma.apiToken.count({ where: { userId: user.id } }),
      prisma.apiToken.count({ where: { userId: user.id, isActive: true, OR: [{ expiresAt: null }, { expiresAt: { gt: now } }] } }),
      prisma.apiToken.count({ where: { userId: user.id, isActive: false } }),
      prisma.apiToken.count({ where: { userId: user.id, expiresAt: { lt: now } } }),
    ]);

    return apiSuccess({ totalTokens, activeTokens, inactiveTokens, expiredTokens });
  } catch (error) {
    console.error("Get token stats error:", error);
    return apiError("获取令牌统计失败");
  }
}
