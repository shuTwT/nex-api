import { NextRequest } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  try {
    const plans = await prisma.subscriptionPlan.findMany({
      where: { isActive: true },
      orderBy: { sortOrder: "asc" },
      select: { id: true, title: true },
    });
    return apiSuccess(plans);
  } catch (error) {
    console.error("Error fetching subscription plans:", error);
    return apiError("获取订阅计划失败");
  }
}
