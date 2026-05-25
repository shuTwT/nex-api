import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";

export async function GET(request: NextRequest) {
  const admin = await getAdminUser();
  if (admin instanceof NextResponse) return admin;

  try {
    const { searchParams } = new URL(request.url);
    const search = searchParams.get("search") || undefined;
    const isActive = searchParams.get("isActive") || undefined;
    const page = parseInt(searchParams.get("page") || "1");
    const limit = parseInt(searchParams.get("limit") || "10");
    const skip = (page - 1) * limit;

    const where: Record<string, unknown> = {};
    if (search) {
      where.title = { contains: search };
    }
    if (isActive !== undefined && isActive !== "") {
      where.isActive = isActive === "true";
    }

    const [plans, total] = await Promise.all([
      prisma.subscriptionPlan.findMany({
        where,
        orderBy: { sortOrder: "asc" },
        skip,
        take: limit,
      }),
      prisma.subscriptionPlan.count({ where }),
    ]);

    return apiPaginated(plans, { page, limit, total });
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
