import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";
import { userUpdateSchema } from "@/lib/validations/user";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  const { id } = await params;

  const found = await prisma.user.findUnique({
    where: { id },
    select: {
      id: true, email: true, username: true, role: true, credits: true,
      createdAt: true, updatedAt: true,
      subscriptions: { where: { isActive: true }, select: { id: true, plan: true, credits: true, price: true, startDate: true, endDate: true, isActive: true }, orderBy: { createdAt: "desc" }, take: 1 },
      apiUsage: { select: { id: true, credits: true, status: true, createdAt: true, api: { select: { name: true, endpoint: true } } }, orderBy: { createdAt: "desc" }, take: 10 },
    },
  });

  if (!found) return apiError("User not found", 404);

  return apiSuccess({ ...found, subscription: found.subscriptions[0] || null, subscriptions: undefined });
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
    const validatedData = userUpdateSchema.parse(body);

    if (validatedData.email || validatedData.username) {
      const existingUser = await prisma.user.findFirst({
        where: {
          AND: [
            { id: { not: id } },
            { OR: [
              ...(validatedData.email ? [{ email: validatedData.email }] : []),
              ...(validatedData.username ? [{ username: validatedData.username }] : []),
            ]},
          ],
        },
      });
      if (existingUser) {
        return apiError(existingUser.email === validatedData.email ? "邮箱已存在" : "用户名已存在", 409);
      }
    }

    const updated = await prisma.user.update({
      where: { id },
      data: validatedData,
      select: { id: true, email: true, username: true, role: true, credits: true, createdAt: true, updatedAt: true },
    });

    revalidatePath("/console/users");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Error updating user:", error);
    return apiError("更新用户失败");
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
    await prisma.user.delete({ where: { id } });
    revalidatePath("/console/users");
    return apiSuccess({ message: "用户已删除" });
  } catch (error) {
    console.error("Error deleting user:", error);
    return apiError("删除用户失败");
  }
}
