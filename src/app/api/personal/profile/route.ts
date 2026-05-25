import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const user = await prisma.user.findUnique({
      where: { id: sessionUser.id },
      select: { id: true, name: true, email: true, image: true, username: true, role: true, credits: true, createdAt: true },
    });

    if (!user) return apiError("用户未找到", 404);

    const [totalCreditsSpent, totalRequests] = await Promise.all([
      prisma.apiUsage.aggregate({ where: { userId: user.id }, _sum: { credits: true } }),
      prisma.apiUsage.count({ where: { userId: user.id } }),
    ]);

    return apiSuccess({ ...user, totalCreditsSpent: totalCreditsSpent._sum.credits || 0, totalRequests });
  } catch (error) {
    console.error("Error fetching user profile:", error);
    return apiError("获取用户信息失败");
  }
}
