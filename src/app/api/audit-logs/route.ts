import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";

export async function GET(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  const { searchParams } = new URL(request.url);
  const search = searchParams.get("search") || undefined;
  const level = searchParams.get("level") || undefined;
  const status = searchParams.get("status") || undefined;
  const startDate = searchParams.get("startDate") || undefined;
  const endDate = searchParams.get("endDate") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (search) {
    where.OR = [
      { action: { contains: search } },
      { resource: { contains: search } },
      { details: { contains: search } },
    ];
  }
  if (level && level !== "all") where.level = level;
  if (status && status !== "all") where.status = status;
  if (startDate || endDate) {
    where.createdAt = {} as Record<string, Date>;
    if (startDate) (where.createdAt as Record<string, Date>).gte = new Date(startDate);
    if (endDate) (where.createdAt as Record<string, Date>).lte = new Date(endDate);
  }

  const [logs, total] = await Promise.all([
    prisma.auditLog.findMany({
      where,
      skip, take: limit,
      orderBy: { createdAt: "desc" },
      include: { user: { select: { id: true, name: true, email: true } } },
    }),
    prisma.auditLog.count({ where }),
  ]);

  return apiPaginated(logs, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { action, resource, details, ipAddress, userAgent, level = "info", status = "success", metadata } = body;

    if (!action || !resource) return apiError("操作和资源不能为空", 400);

    const newLog = await prisma.auditLog.create({
      data: { userId: user.id, action, resource, details: details || null, ipAddress: ipAddress || null, userAgent: userAgent || null, level, status, metadata: metadata || null },
    });

    revalidatePath("/console/audit-logs");
    return apiSuccess(newLog, 201);
  } catch (error) {
    console.error("Create audit log error:", error);
    return apiError("创建审计日志失败");
  }
}
