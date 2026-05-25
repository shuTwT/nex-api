import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  const { id } = await params;
  const api = await prisma.api.findUnique({
    where: { id },
    include: { category: true, parameters: true, responses: true },
  });

  if (!api) return apiError("API 不存在", 404);
  return apiSuccess(api);
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
    const { name, alias, description, endpoint, method, categoryId, pricing, documentation, preScript, postScript, isActive } = body;

    if (!name || !alias || !endpoint || !method || !categoryId) {
      return apiError("缺少必填字段", 400);
    }

    const aliasPattern = /^[a-zA-Z][a-zA-Z0-9]*$/;
    if (!aliasPattern.test(alias)) return apiError("别名必须以字母开头，只能包含字母和数字", 400);

    const existingAlias = await prisma.api.findFirst({ where: { alias, NOT: { id } } });
    if (existingAlias) return apiError("别名已存在", 409);

    const existingEndpoint = await prisma.api.findFirst({ where: { endpoint, NOT: { id } } });
    if (existingEndpoint) return apiError("接口端点已存在", 409);

    const updated = await prisma.api.update({
      where: { id },
      data: { name, alias, description, endpoint, method, categoryId, pricing: pricing || 0, documentation, preScript, postScript, isActive },
      include: { category: true },
    });

    await logAudit({ action: "更新 API", resource: "API 管理", details: `更新了 API: ${name} (${alias})`, level: "info", status: "success" });

    revalidatePath("/console/api-management");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Update API error:", error);
    await logAudit({ action: "更新 API", resource: "API 管理", details: `更新 API 失败: ${error}`, level: "error", status: "error" });
    return apiError("更新 API 失败");
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
    const apiToDelete = await prisma.api.findUnique({ where: { id } });
    await prisma.api.delete({ where: { id } });

    await logAudit({ action: "删除 API", resource: "API 管理", details: `删除了 API: ${apiToDelete?.name || id}`, level: "warning", status: "success" });

    revalidatePath("/console/api-management");
    return apiSuccess({ message: "API 已删除" });
  } catch (error) {
    console.error("Delete API error:", error);
    await logAudit({ action: "删除 API", resource: "API 管理", details: `删除 API 失败: ${error}`, level: "error", status: "error" });
    return apiError("删除 API 失败");
  }
}
