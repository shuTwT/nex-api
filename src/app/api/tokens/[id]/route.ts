import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();
    const { name, permissions = "read", expiresAt, isActive } = body;

    if (!name) return apiError("缺少必填字段", 400);

    const existingToken = await prisma.apiToken.findFirst({ where: { userId: user.id, name, NOT: { id } } });
    if (existingToken) return apiError("令牌名称已存在", 409);

    const updated = await prisma.apiToken.update({
      where: { id, userId: user.id },
      data: { name, permissions, expiresAt: expiresAt ? new Date(expiresAt) : null, isActive },
    });

    await logAudit({ userId: user.id, action: "更新令牌", resource: "令牌管理", details: `更新了令牌: ${name}`, level: "info", status: "success" });

    revalidatePath("/console/tokens");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Update token error:", error);
    await logAudit({ action: "更新令牌", resource: "令牌管理", details: `更新令牌失败: ${error}`, level: "error", status: "error" });
    return apiError("更新令牌失败");
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const tokenToDelete = await prisma.apiToken.findUnique({ where: { id, userId: user.id } });
    await prisma.apiToken.delete({ where: { id, userId: user.id } });

    await logAudit({ userId: user.id, action: "删除令牌", resource: "令牌管理", details: `删除了令牌: ${tokenToDelete?.name || id}`, level: "warning", status: "success" });

    revalidatePath("/console/tokens");
    return apiSuccess({ message: "令牌已删除" });
  } catch (error) {
    console.error("Delete token error:", error);
    await logAudit({ action: "删除令牌", resource: "令牌管理", details: `删除令牌失败: ${error}`, level: "error", status: "error" });
    return apiError("删除令牌失败");
  }
}
