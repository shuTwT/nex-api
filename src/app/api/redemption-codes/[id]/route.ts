import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const code = await prisma.redemptionCode.findUnique({ where: { id } });
    if (!code) return apiError("兑换码未找到", 404);
    if (code.isUsed) return apiError("该兑换码已被使用，无法删除", 400);

    await prisma.redemptionCode.delete({ where: { id } });

    revalidatePath("/console/redemption-codes");
    return apiSuccess({ message: "兑换码已删除" });
  } catch (error) {
    console.error("Error deleting redemption code:", error);
    return apiError("删除兑换码失败");
  }
}
