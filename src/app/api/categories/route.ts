import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const categories = await prisma.apiCategory.findMany({
      include: { _count: { select: { apis: true } } },
      orderBy: { name: "asc" },
    });

    return apiSuccess(categories.map((cat) => ({
      id: cat.id, name: cat.name, description: cat.description, icon: cat.icon, apiCount: cat._count.apis,
    })));
  } catch (error) {
    console.error("Get categories error:", error);
    return apiError("获取分类列表失败");
  }
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { name, description = "", icon } = body;

    if (!name) return apiError("分类名称不能为空", 400);

    const existingCategory = await prisma.apiCategory.findUnique({ where: { name } });
    if (existingCategory) return apiError("分类名称已存在", 409);

    const category = await prisma.apiCategory.create({ data: { name, description, icon } });

    revalidatePath("/console/api-management");
    return apiSuccess(category, 201);
  } catch (error) {
    console.error("Create category error:", error);
    return apiError("创建分类失败");
  }
}
