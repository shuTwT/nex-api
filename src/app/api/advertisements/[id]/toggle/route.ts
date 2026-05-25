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
    const advertisement = await prisma.advertisement.findUnique({ where: { id }, select: { isActive: true } });
    if (!advertisement) return apiError("广告不存在", 404);

    const updated = await prisma.advertisement.update({ where: { id }, data: { isActive: !advertisement.isActive } });

    revalidatePath("/console/advertisements");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Error toggling advertisement status:", error);
    return apiError("切换广告状态失败");
  }
}
