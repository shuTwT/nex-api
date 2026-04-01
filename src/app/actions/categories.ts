"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/auth/util";
import { authOptions } from "@/lib/auth/config";

export async function getCategories() {
  try {
    await requireAdmin(authOptions);

    const categories = await prisma.apiCategory.findMany({
      include: {
        _count: {
          select: {
            apis: true,
          },
        },
      },
      orderBy: {
        name: "asc",
      },
    });

    return {
      success: true,
      data: categories.map((cat) => ({
        id: cat.id,
        name: cat.name,
        description: cat.description,
        icon: cat.icon,
        apiCount: cat._count.apis,
      })),
    };
  } catch (error) {
    console.error("Get categories error:", error);
    return {
      success: false,
      error: "获取分类列表失败",
    };
  }
}

export async function createCategory(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const name = formData.get("name") as string;
    const description = formData.get("description") as string;
    const icon = formData.get("icon") as string;

    if (!name) {
      return {
        success: false,
        error: "分类名称不能为空",
      };
    }

    const existingCategory = await prisma.apiCategory.findUnique({
      where: { name },
    });

    if (existingCategory) {
      return {
        success: false,
        error: "分类名称已存在",
      };
    }

    const category = await prisma.apiCategory.create({
      data: {
        name,
        description: description || "",
        icon,
      },
    });

    revalidatePath("/console/api-management");
    return {
      success: true,
      data: category,
    };
  } catch (error) {
    console.error("Create category error:", error);
    return {
      success: false,
      error: "创建分类失败",
    };
  }
}

export async function updateCategory(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const id = formData.get("id") as string;
    const name = formData.get("name") as string;
    const description = formData.get("description") as string;
    const icon = formData.get("icon") as string;

    if (!id || !name) {
      return {
        success: false,
        error: "缺少必填字段",
      };
    }

    const existingCategory = await prisma.apiCategory.findFirst({
      where: {
        name,
        NOT: { id },
      },
    });

    if (existingCategory) {
      return {
        success: false,
        error: "分类名称已存在",
      };
    }

    const category = await prisma.apiCategory.update({
      where: { id },
      data: {
        name,
        description: description || "",
        icon,
      },
    });

    revalidatePath("/console/api-management");
    return {
      success: true,
      data: category,
    };
  } catch (error) {
    console.error("Update category error:", error);
    return {
      success: false,
      error: "更新分类失败",
    };
  }
}

export async function deleteCategory(id: string) {
  try {
    await requireAdmin(authOptions);

    const apisCount = await prisma.api.count({
      where: { categoryId: id },
    });

    if (apisCount > 0) {
      return {
        success: false,
        error: "该分类下还有 API，无法删除",
      };
    }

    await prisma.apiCategory.delete({
      where: { id },
    });

    revalidatePath("/console/api-management");
    return {
      success: true,
    };
  } catch (error) {
    console.error("Delete category error:", error);
    return {
      success: false,
      error: "删除分类失败",
    };
  }
}
