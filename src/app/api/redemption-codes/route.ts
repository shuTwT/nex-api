import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import crypto from "crypto";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";

function generateCode(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let code = "";
  const bytes = crypto.randomBytes(12);
  for (let i = 0; i < 12; i++) code += chars[bytes[i] % chars.length];
  return code;
}

export async function GET(request: NextRequest) {
  const admin = await getAdminUser();
  if (admin instanceof NextResponse) return admin;

  const { searchParams } = new URL(request.url);
  const search = searchParams.get("search") || undefined;
  const type = searchParams.get("type") || undefined;
  const isUsedParam = searchParams.get("isUsed") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (search) {
    where.code = { contains: search };
  }
  if (type && type !== "all") {
    where.type = type;
  }
  if (isUsedParam !== undefined && isUsedParam !== "") {
    where.isUsed = isUsedParam === "true";
  }

  const [codes, total] = await Promise.all([
    prisma.redemptionCode.findMany({ where, orderBy: { createdAt: "desc" }, skip, take: limit }),
    prisma.redemptionCode.count({ where }),
  ]);

  return apiPaginated(codes, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const sessionUser = await getAdminUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const body = await request.json();
    const { type, count = 1, planId = null, credits = null, expiresAt: expiresAtStr = null } = body;

    if (!type || (type !== "subscription" && type !== "quota")) return apiError("请选择有效的兑换码类型", 400);
    if (type === "subscription" && !planId) return apiError("请选择订阅计划", 400);
    if (type === "quota" && (!credits || credits <= 0)) return apiError("请输入有效额度", 400);
    if (count < 1 || count > 1000) return apiError("生成数量范围为 1-1000", 400);

    const batchId = crypto.randomUUID();
    const codes: string[] = [];
    for (let i = 0; i < count; i++) {
      let code: string;
      let attempts = 0;
      do {
        code = generateCode();
        attempts++;
        if (attempts > 10) return apiError("生成兑换码失败，请重试");
      } while (codes.includes(code));
      codes.push(code);
    }

    const existingCodes = await prisma.redemptionCode.findMany({
      where: { code: { in: codes } }, select: { code: true },
    });
    if (existingCodes.length > 0) return apiError("生成的兑换码与已有兑换码冲突，请重试", 409);

    let planName: string | null = null;
    if (type === "subscription" && planId) {
      const plan = await prisma.subscriptionPlan.findUnique({ where: { id: planId }, select: { title: true } });
      planName = plan?.title ?? null;
    }

    const expiresAt = expiresAtStr ? new Date(expiresAtStr) : null;

    await prisma.redemptionCode.createMany({
      data: codes.map((code) => ({
        code, type, planId: type === "subscription" ? planId : null, planName,
        credits: type === "quota" ? credits : null, expiresAt, createdBy: sessionUser.id, batchId,
      })),
    });

    revalidatePath("/console/redemption-codes");
    return apiSuccess({ count, batchId }, 201);
  } catch (error) {
    console.error("Error creating redemption codes:", error);
    return apiError("创建兑换码失败");
  }
}
