import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const plan = await prisma.subscriptionPlan.findUnique({ where: { id } });
  if (!plan) return apiError("订阅计划未找到", 404);
  return apiSuccess(plan);
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();

    const data: Record<string, unknown> = {};
    if (body.title !== undefined) data.title = body.title;
    if (body.price !== undefined) data.price = body.price;
    if (body.totalCredits !== undefined) data.totalCredits = body.totalCredits;
    if (body.sortOrder !== undefined) data.sortOrder = body.sortOrder;
    if (body.validityDuration !== undefined) data.validityDuration = body.validityDuration;
    if (body.validityUnit !== undefined) data.validityUnit = body.validityUnit;
    if (body.creditResetCycle !== undefined) data.creditResetCycle = body.creditResetCycle;
    if (body.isActive !== undefined) data.isActive = body.isActive;

    const plan = await prisma.subscriptionPlan.update({ where: { id }, data });

    revalidatePath("/console/subscription-plans");
    return apiSuccess(plan);
  } catch (error) {
    console.error("Error updating subscription plan:", error);
    return apiError("更新订阅计划失败");
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    await prisma.subscriptionPlan.delete({ where: { id } });

    revalidatePath("/console/subscription-plans");
    return apiSuccess({ message: "订阅计划已删除" });
  } catch (error) {
    console.error("Error deleting subscription plan:", error);
    return apiError("删除订阅计划失败");
  }
}
