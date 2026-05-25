import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function PUT(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const api = await prisma.api.findUnique({ where: { id } });
    if (!api) return apiError("API 不存在", 404);

    const updated = await prisma.api.update({ where: { id }, data: { isActive: !api.isActive } });
    revalidatePath("/console/api-management");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Toggle API status error:", error);
    return apiError("切换 API 状态失败");
  }
}
