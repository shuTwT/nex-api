import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function POST(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { code: codeInput } = body;

    const trimmed = codeInput?.trim().toUpperCase();
    if (!trimmed) return apiError("请输入兑换码", 400);

    const code = await prisma.redemptionCode.findUnique({ where: { code: trimmed } });
    if (!code) return apiError("兑换码不存在", 404);
    if (code.isUsed) return apiError("该兑换码已被使用", 400);
    if (code.expiresAt && new Date(code.expiresAt) < new Date()) return apiError("该兑换码已过期", 400);

    return apiSuccess({ type: code.type, planName: code.planName, credits: code.credits });
  } catch (error) {
    console.error("Error looking up redemption code:", error);
    return apiError("查询兑换码失败");
  }
}
