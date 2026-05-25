import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function DELETE(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { searchParams } = new URL(request.url);
    const batchId = searchParams.get("batchId");

    if (!batchId) return apiError("缺少批次 ID", 400);

    const usedCount = await prisma.redemptionCode.count({ where: { batchId, isUsed: true } });
    if (usedCount > 0) return apiError(`该批次有 ${usedCount} 个兑换码已被使用，无法批量删除`, 400);

    const result = await prisma.redemptionCode.deleteMany({ where: { batchId, isUsed: false } });

    revalidatePath("/console/redemption-codes");
    return apiSuccess({ message: `已删除 ${result.count} 个兑换码` });
  } catch (error) {
    console.error("Error batch deleting redemption codes:", error);
    return apiError("批量删除兑换码失败");
  }
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { ids } = body;

    if (!ids || ids.length === 0) return apiError("请选择要删除的兑换码", 400);

    const usedCount = await prisma.redemptionCode.count({ where: { id: { in: ids }, isUsed: true } });
    if (usedCount > 0) return apiError(`所选中有 ${usedCount} 个兑换码已被使用，无法删除`, 400);

    const result = await prisma.redemptionCode.deleteMany({ where: { id: { in: ids }, isUsed: false } });

    revalidatePath("/console/redemption-codes");
    return apiSuccess({ message: `已删除 ${result.count} 个兑换码` });
  } catch (error) {
    console.error("Error batch deleting redemption codes:", error);
    return apiError("批量删除兑换码失败");
  }
}
