import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function PUT(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const token = await prisma.apiToken.findUnique({ where: { id, userId: user.id } });
    if (!token) return apiError("令牌不存在", 404);

    const updated = await prisma.apiToken.update({ where: { id }, data: { isActive: !token.isActive } });
    revalidatePath("/console/tokens");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Toggle token status error:", error);
    return apiError("切换令牌状态失败");
  }
}
