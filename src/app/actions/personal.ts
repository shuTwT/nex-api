"use server";

import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/auth/util";
import { authOptions } from "@/lib/auth/config";

export async function getCurrentUserProfile() {
  try {
    const sessionUser = await requireAuth(authOptions);

    console.log(sessionUser)
    
    const user = await prisma.user.findUnique({
      where: { id: sessionUser.id },
      select: {
        id: true,
        name: true,
        email: true,
        image: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
      },
    });

    if (!user) {
      return { success: false, error: "用户未找到" };
    }

    const totalCreditsSpent = await prisma.apiUsage.aggregate({
      where: {
        userId: user.id,
      },
      _sum: {
        credits: true,
      },
    });

    const totalRequests = await prisma.apiUsage.count({
      where: {
        userId: user.id,
      },
    });

    return {
      success: true,
      data: {
        ...user,
        totalCreditsSpent: totalCreditsSpent._sum.credits || 0,
        totalRequests,
      },
    };
  } catch (error) {
    console.error("Error fetching user profile:", error);
    return { success: false, error: "获取用户信息失败" };
  }
}
