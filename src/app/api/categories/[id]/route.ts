import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();
    const { name, description = "", icon } = body;

    if (!name) return apiError("缺少必填字段", 400);

    const existingCategory = await prisma.apiCategory.findFirst({ where: { name, NOT: { id } } });
    if (existingCategory) return apiError("分类名称已存在", 409);

    const category = await prisma.apiCategory.update({ where: { id }, data: { name, description, icon } });

    revalidatePath("/console/api-management");
    return apiSuccess(category);
  } catch (error) {
    console.error("Update category error:", error);
    return apiError("更新分类失败");
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
    const apisCount = await prisma.api.count({ where: { categoryId: id } });
    if (apisCount > 0) return apiError("该分类下还有 API，无法删除", 400);

    await prisma.apiCategory.delete({ where: { id } });

    revalidatePath("/console/api-management");
    return apiSuccess({ message: "分类已删除" });
  } catch (error) {
    console.error("Delete category error:", error);
    return apiError("删除分类失败");
  }
}
