"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/auth/util";
import { logAudit } from "@/lib/audit-log";
import { authOptions } from "@/lib/auth/config";

export async function getApis(params: {
  category?: string;
  search?: string;
  status?: string;
  page?: number;
  limit?: number;
}): Promise<{
  success: boolean;
  data?: any[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  error?: string;
}> {
  try {
    await requireAdmin(authOptions);

    const { category, search, status, page = 1, limit = 10 } = params;
    const skip = (page - 1) * limit;

    const where: any = {};

    if (category && category !== "all") {
      where.categoryId = category;
    }

    if (search) {
      where.OR = [
        { name: { contains: search } },
        { description: { contains: search } },
        { endpoint: { contains: search } },
      ];
    }

    if (status === "active") {
      where.isActive = true;
    } else if (status === "inactive") {
      where.isActive = false;
    }

    const [apis, total] = await Promise.all([
      prisma.api.findMany({
        where,
        skip,
        take: limit,
        include: {
          category: {
            select: {
              id: true,
              name: true,
            },
          },
        },
        orderBy: {
          createdAt: "desc",
        },
      }),
      prisma.api.count({ where }),
    ]);

    return {
      success: true,
      data: apis,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Get APIs error:", error);
    return {
      success: false,
      error: "获取 API 列表失败",
    };
  }
}

export async function getApiById(id: string) {
  try {
    await requireAdmin(authOptions);

    const api = await prisma.api.findUnique({
      where: { id },
      include: {
        category: true,
        parameters: true,
        responses: true,
      },
    });

    if (!api) {
      return { success: false, error: "API 不存在" };
    }

    return { success: true, data: api };
  } catch (error) {
    console.error("Get API error:", error);
    return { success: false, error: "获取 API 详情失败" };
  }
}

export async function createApi(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const name = formData.get("name") as string;
    const alias = formData.get("alias") as string;
    const description = formData.get("description") as string;
    const endpoint = formData.get("endpoint") as string;
    const method = formData.get("method") as string;
    const categoryId = formData.get("categoryId") as string;
    const pricingStr = formData.get("pricing") as string;
    const pricing = parseInt(pricingStr) || 0;
    const documentation = formData.get("documentation") as string;
    const preScript = formData.get("preScript") as string;
    const postScript = formData.get("postScript") as string;
    const isActive = formData.get("isActive") === "true";

    if (!name || !alias || !endpoint || !method || !categoryId) {
      return { success: false, error: "缺少必填字段" };
    }

    const aliasPattern = /^[a-zA-Z][a-zA-Z0-9]*$/;
    if (!aliasPattern.test(alias)) {
      return { success: false, error: "别名必须以字母开头，只能包含字母和数字" };
    }

    const existingAlias = await prisma.api.findUnique({
      where: { alias },
    });

    if (existingAlias) {
      return { success: false, error: "别名已存在" };
    }

    const existingApi = await prisma.api.findUnique({
      where: { endpoint },
    });

    if (existingApi) {
      return { success: false, error: "接口端点已存在" };
    }

    const api = await prisma.api.create({
      data: {
        name,
        alias,
        description,
        endpoint,
        method,
        categoryId,
        pricing,
        documentation,
        preScript,
        postScript,
        isActive,
      },
      include: {
        category: true,
      },
    });

    await logAudit({
      action: "创建 API",
      resource: "API 管理",
      details: `创建了 API: ${name} (${alias})`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/api-management");
    return { success: true, data: api };
  } catch (error) {
    console.error("Create API error:", error);
    await logAudit({
      action: "创建 API",
      resource: "API 管理",
      details: `创建 API 失败: ${error}`,
      level: "error",
      status: "error",
    });
    return { success: false, error: "创建 API 失败" };
  }
}

export async function updateApi(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const id = formData.get("id") as string;
    const name = formData.get("name") as string;
    const alias = formData.get("alias") as string;
    const description = formData.get("description") as string;
    const endpoint = formData.get("endpoint") as string;
    const method = formData.get("method") as string;
    const categoryId = formData.get("categoryId") as string;
    const pricingStr = formData.get("pricing") as string;
    const pricing = parseInt(pricingStr) || 0;
    const documentation = formData.get("documentation") as string;
    const preScript = formData.get("preScript") as string;
    const postScript = formData.get("postScript") as string;
    const isActive = formData.get("isActive") === "true";

    if (!id || !name || !alias || !endpoint || !method || !categoryId) {
      return { success: false, error: "缺少必填字段" };
    }

    const aliasPattern = /^[a-zA-Z][a-zA-Z0-9]*$/;
    if (!aliasPattern.test(alias)) {
      return { success: false, error: "别名必须以字母开头，只能包含字母和数字" };
    }

    const existingAlias = await prisma.api.findFirst({
      where: {
        alias,
        NOT: { id },
      },
    });

    if (existingAlias) {
      return { success: false, error: "别名已存在" };
    }

    const existingApi = await prisma.api.findFirst({
      where: {
        endpoint,
        NOT: { id },
      },
    });

    if (existingApi) {
      return { success: false, error: "接口端点已存在" };
    }

    const api = await prisma.api.update({
      where: { id },
      data: {
        name,
        alias,
        description,
        endpoint,
        method,
        categoryId,
        pricing: pricing || 0,
        documentation,
        preScript,
        postScript,
        isActive,
      },
      include: {
        category: true,
      },
    });

    await logAudit({
      action: "更新 API",
      resource: "API 管理",
      details: `更新了 API: ${name} (${alias})`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/api-management");
    return { success: true, data: api };
  } catch (error) {
    console.error("Update API error:", error);
    await logAudit({
      action: "更新 API",
      resource: "API 管理",
      details: `更新 API 失败: ${error}`,
      level: "error",
      status: "error",
    });
    return { success: false, error: "更新 API 失败" };
  }
}

export async function deleteApi(id: string) {
  try {
    await requireAdmin(authOptions);

    const apiToDelete = await prisma.api.findUnique({
      where: { id },
    });

    await prisma.api.delete({
      where: { id },
    });

    await logAudit({
      action: "删除 API",
      resource: "API 管理",
      details: `删除了 API: ${apiToDelete?.name || id}`,
      level: "warning",
      status: "success",
    });

    revalidatePath("/console/api-management");
    return { success: true };
  } catch (error) {
    console.error("Delete API error:", error);
    await logAudit({
      action: "删除 API",
      resource: "API 管理",
      details: `删除 API 失败: ${error}`,
      level: "error",
      status: "error",
    });
    return { success: false, error: "删除 API 失败" };
  }
}

export async function toggleApiStatus(id: string) {
  try {
    await requireAdmin(authOptions);

    const api = await prisma.api.findUnique({
      where: { id },
    });

    if (!api) {
      return { success: false, error: "API 不存在" };
    }

    const updatedApi = await prisma.api.update({
      where: { id },
      data: {
        isActive: !api.isActive,
      },
    });

    revalidatePath("/console/api-management");
    return { success: true, data: updatedApi };
  } catch (error) {
    console.error("Toggle API status error:", error);
    return { success: false, error: "切换 API 状态失败" };
  }
}

export async function getApiStats() {
  try {
    await requireAdmin(authOptions);

    const [totalApis, activeApis, inactiveApis, totalCalls, categoriesCount] = await Promise.all([
      prisma.api.count(),
      prisma.api.count({ where: { isActive: true } }),
      prisma.api.count({ where: { isActive: false } }),
      prisma.api.aggregate({
        _sum: {
          callCount: true,
        },
      }),
      prisma.apiCategory.count(),
    ]);

    return {
      success: true,
      data: {
        totalApis,
        activeApis,
        inactiveApis,
        totalCalls: totalCalls._sum.callCount || 0,
        categoriesCount,
      },
    };
  } catch (error) {
    console.error("Get API stats error:", error);
    return { success: false, error: "获取 API 统计失败" };
  }
}
