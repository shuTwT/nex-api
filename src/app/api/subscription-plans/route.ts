import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  try {
    const plans = await prisma.subscriptionPlan.findMany({ orderBy: { sortOrder: "asc" } });
    return apiSuccess(plans);
  } catch (error) {
    console.error("Error fetching subscription plans:", error);
    return apiError("获取订阅计划失败");
  }
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { title, price = 0, totalCredits = 0, sortOrder = 0, validityDuration = 30, validityUnit = "day", creditResetCycle = "month", isActive = false } = body;

    const plan = await prisma.subscriptionPlan.create({
      data: { title, price, totalCredits, sortOrder, validityDuration, validityUnit, creditResetCycle, isActive },
    });

    revalidatePath("/console/subscription-plans");
    return apiSuccess(plan, 201);
  } catch (error) {
    console.error("Error creating subscription plan:", error);
    return apiError("创建订阅计划失败");
  }
}
