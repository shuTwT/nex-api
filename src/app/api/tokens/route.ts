import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { randomBytes } from "crypto";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

function generateToken(): string {
  return `sk_${randomBytes(32).toString("hex")}`;
}

export async function GET(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  const { searchParams } = new URL(request.url);
  const search = searchParams.get("search") || undefined;
  const status = searchParams.get("status") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = { userId: user.id };
  if (search) where.name = { contains: search };
  if (status === "active") where.isActive = true;
  else if (status === "inactive") where.isActive = false;

  const [tokens, total] = await Promise.all([
    prisma.apiToken.findMany({ where, skip, take: limit, orderBy: { createdAt: "desc" } }),
    prisma.apiToken.count({ where }),
  ]);

  return apiPaginated(tokens, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { name, permissions = "read", expiresAt } = body;

    if (!name) return apiError("令牌名称不能为空", 400);

    const existingToken = await prisma.apiToken.findFirst({ where: { userId: user.id, name } });
    if (existingToken) return apiError("令牌名称已存在", 409);

    const token = generateToken();
    const newToken = await prisma.apiToken.create({
      data: { userId: user.id, name, token, permissions, expiresAt: expiresAt ? new Date(expiresAt) : null },
    });

    await logAudit({ userId: user.id, action: "创建令牌", resource: "令牌管理", details: `创建了令牌: ${name}`, level: "info", status: "success" });

    revalidatePath("/console/tokens");
    return apiSuccess(newToken, 201);
  } catch (error) {
    console.error("Create token error:", error);
    await logAudit({ action: "创建令牌", resource: "令牌管理", details: `创建令牌失败: ${error}`, level: "error", status: "error" });
    return apiError("创建令牌失败");
  }
}
