import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();
    const { action, resource, details, ipAddress, userAgent, level = "info", status = "success", metadata } = body;

    if (!action || !resource) return apiError("缺少必填字段", 400);

    const updated = await prisma.auditLog.update({
      where: { id },
      data: { action, resource, details: details || null, ipAddress: ipAddress || null, userAgent: userAgent || null, level, status, metadata: metadata || null },
    });

    revalidatePath("/console/audit-logs");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Update audit log error:", error);
    return apiError("更新审计日志失败");
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
    await prisma.auditLog.delete({ where: { id } });
    revalidatePath("/console/audit-logs");
    return apiSuccess({ message: "审计日志已删除" });
  } catch (error) {
    console.error("Delete audit log error:", error);
    return apiError("删除审计日志失败");
  }
}
