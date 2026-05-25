import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const subscription = await prisma.subscription.findFirst({
      where: { userId: sessionUser.id, isActive: true, endDate: { gte: new Date() } },
      include: { plan: true },
      orderBy: { createdAt: "desc" },
    });

    return apiSuccess(subscription);
  } catch (error) {
    console.error("Error fetching current subscription:", error);
    return apiError("获取订阅信息失败");
  }
}
