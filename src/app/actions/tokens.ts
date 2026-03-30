"use server";

import { revalidatePath } from "next/cache";
import { randomBytes } from "crypto";
import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/session";
import { logAudit } from "@/lib/audit-log";

function generateToken(): string {
  const prefix = "sk";
  const randomString = randomBytes(32).toString("hex");
  return `${prefix}_${randomString}`;
}

export async function getTokens(params?: {
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
    const user = await requireAuth();
    const { search, status, page = 1, limit = 10 } = params || {};
    const skip = (page - 1) * limit;

    const where: any = {
      userId: user.id,
    };

    if (search) {
      where.name = { contains: search };
    }

    if (status === "active") {
      where.isActive = true;
    } else if (status === "inactive") {
      where.isActive = false;
    }

    const [tokens, total] = await Promise.all([
      prisma.apiToken.findMany({
        where,
        skip,
        take: limit,
        orderBy: {
          createdAt: "desc",
        },
      }),
      prisma.apiToken.count({ where }),
    ]);

    return {
      success: true,
      data: tokens,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Get tokens error:", error);
    return {
      success: false,
      error: "获取令牌列表失败",
    };
  }
}

export async function createToken(formData: FormData): Promise<{
  success: boolean;
  data?: any;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const name = formData.get("name") as string;
    const permissions = formData.get("permissions") as string;
    const expiresAt = formData.get("expiresAt") as string;

    if (!name) {
      return {
        success: false,
        error: "令牌名称不能为空",
      };
    }

    const existingToken = await prisma.apiToken.findFirst({
      where: {
        userId: user.id,
        name,
      },
    });

    if (existingToken) {
      return {
        success: false,
        error: "令牌名称已存在",
      };
    }

    const token = generateToken();

    console.log(user);

    const newToken = await prisma.apiToken.create({
      data: {
        userId: user.id,
        name,
        token,
        permissions: permissions || "read",
        expiresAt: expiresAt ? new Date(expiresAt) : null,
      },
    });

    await logAudit({
      userId: user.id,
      action: "创建令牌",
      resource: "令牌管理",
      details: `创建了令牌: ${name}`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/tokens");
    return {
      success: true,
      data: newToken,
    };
  } catch (error) {
    console.error("Create token error:", error);
    await logAudit({
      action: "创建令牌",
      resource: "令牌管理",
      details: `创建令牌失败: ${error}`,
      level: "error",
      status: "error",
    });
    return {
      success: false,
      error: "创建令牌失败",
    };
  }
}

export async function updateToken(formData: FormData): Promise<{
  success: boolean;
  data?: any;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const id = formData.get("id") as string;
    const name = formData.get("name") as string;
    const permissions = formData.get("permissions") as string;
    const expiresAt = formData.get("expiresAt") as string;
    const isActive = formData.get("isActive") === "true";

    if (!id || !name) {
      return {
        success: false,
        error: "缺少必填字段",
      };
    }

    const existingToken = await prisma.apiToken.findFirst({
      where: {
        userId: user.id,
        name,
        NOT: { id },
      },
    });

    if (existingToken) {
      return {
        success: false,
        error: "令牌名称已存在",
      };
    }

    const updatedToken = await prisma.apiToken.update({
      where: {
        id,
        userId: user.id,
      },
      data: {
        name,
        permissions: permissions || "read",
        expiresAt: expiresAt ? new Date(expiresAt) : null,
        isActive,
      },
    });

    await logAudit({
      userId: user.id,
      action: "更新令牌",
      resource: "令牌管理",
      details: `更新了令牌: ${name}`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/tokens");
    return {
      success: true,
      data: updatedToken,
    };
  } catch (error) {
    console.error("Update token error:", error);
    await logAudit({
      action: "更新令牌",
      resource: "令牌管理",
      details: `更新令牌失败: ${error}`,
      level: "error",
      status: "error",
    });
    return {
      success: false,
      error: "更新令牌失败",
    };
  }
}

export async function deleteToken(id: string): Promise<{
  success: boolean;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const tokenToDelete = await prisma.apiToken.findUnique({
      where: {
        id,
        userId: user.id,
      },
    });

    await prisma.apiToken.delete({
      where: {
        id,
        userId: user.id,
      },
    });

    await logAudit({
      userId: user.id,
      action: "删除令牌",
      resource: "令牌管理",
      details: `删除了令牌: ${tokenToDelete?.name || id}`,
      level: "warning",
      status: "success",
    });

    revalidatePath("/console/tokens");
    return {
      success: true,
    };
  } catch (error) {
    console.error("Delete token error:", error);
    await logAudit({
      action: "删除令牌",
      resource: "令牌管理",
      details: `删除令牌失败: ${error}`,
      level: "error",
      status: "error",
    });
    return {
      success: false,
      error: "删除令牌失败",
    };
  }
}

export async function toggleTokenStatus(id: string): Promise<{
  success: boolean;
  data?: any;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const token = await prisma.apiToken.findUnique({
      where: {
        id,
        userId: user.id,
      },
    });

    if (!token) {
      return {
        success: false,
        error: "令牌不存在",
      };
    }

    const updatedToken = await prisma.apiToken.update({
      where: { id },
      data: {
        isActive: !token.isActive,
      },
    });

    revalidatePath("/console/tokens");
    return {
      success: true,
      data: updatedToken,
    };
  } catch (error) {
    console.error("Toggle token status error:", error);
    return {
      success: false,
      error: "切换令牌状态失败",
    };
  }
}

export async function getTokenStats(): Promise<{
  success: boolean;
  data?: {
    totalTokens: number;
    activeTokens: number;
    inactiveTokens: number;
    expiredTokens: number;
  };
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const now = new Date();

    const [totalTokens, activeTokens, inactiveTokens, expiredTokens] = await Promise.all([
      prisma.apiToken.count({
        where: { userId: user.id },
      }),
      prisma.apiToken.count({
        where: {
          userId: user.id,
          isActive: true,
          OR: [
            { expiresAt: null },
            { expiresAt: { gt: now } },
          ],
        },
      }),
      prisma.apiToken.count({
        where: {
          userId: user.id,
          isActive: false,
        },
      }),
      prisma.apiToken.count({
        where: {
          userId: user.id,
          expiresAt: { lt: now },
        },
      }),
    ]);

    return {
      success: true,
      data: {
        totalTokens,
        activeTokens,
        inactiveTokens,
        expiredTokens,
      },
    };
  } catch (error) {
    console.error("Get token stats error:", error);
    return {
      success: false,
      error: "获取令牌统计失败",
    };
  }
}
