import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

export async function GET(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  const { searchParams } = new URL(request.url);
  const category = searchParams.get("category") || undefined;
  const search = searchParams.get("search") || undefined;
  const status = searchParams.get("status") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (category && category !== "all") where.categoryId = category;
  if (search) {
    where.OR = [
      { name: { contains: search } },
      { description: { contains: search } },
      { endpoint: { contains: search } },
    ];
  }
  if (status === "active") where.isActive = true;
  else if (status === "inactive") where.isActive = false;

  const [apis, total] = await Promise.all([
    prisma.api.findMany({
      where,
      skip, take: limit,
      include: { category: { select: { id: true, name: true } } },
      orderBy: { createdAt: "desc" },
    }),
    prisma.api.count({ where }),
  ]);

  return apiPaginated(apis, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { name, alias, description, endpoint, method, categoryId, pricing = 0, documentation, preScript, postScript, isActive = true } = body;

    if (!name || !alias || !endpoint || !method || !categoryId) {
      return apiError("缺少必填字段", 400);
    }

    const aliasPattern = /^[a-zA-Z][a-zA-Z0-9]*$/;
    if (!aliasPattern.test(alias)) {
      return apiError("别名必须以字母开头，只能包含字母和数字", 400);
    }

    const existingAlias = await prisma.api.findUnique({ where: { alias } });
    if (existingAlias) return apiError("别名已存在", 409);

    const existingApi = await prisma.api.findUnique({ where: { endpoint } });
    if (existingApi) return apiError("接口端点已存在", 409);

    const api = await prisma.api.create({
      data: { name, alias, description, endpoint, method, categoryId, pricing, documentation, preScript, postScript, isActive },
      include: { category: true },
    });

    await logAudit({ action: "创建 API", resource: "API 管理", details: `创建了 API: ${name} (${alias})`, level: "info", status: "success" });

    revalidatePath("/console/api-management");
    return apiSuccess(api, 201);
  } catch (error) {
    console.error("Create API error:", error);
    await logAudit({ action: "创建 API", resource: "API 管理", details: `创建 API 失败: ${error}`, level: "error", status: "error" });
    return apiError("创建 API 失败");
  }
}
