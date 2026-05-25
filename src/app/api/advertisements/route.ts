import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";

export async function GET(request: NextRequest) {
  const admin = await getAdminUser();
  if (admin instanceof NextResponse) return admin;

  const { searchParams } = new URL(request.url);
  const search = searchParams.get("search") || undefined;
  const position = searchParams.get("position") || undefined;
  const isActiveParam = searchParams.get("isActive");
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (search) {
    where.title = { contains: search };
  }
  if (position) where.position = position;
  if (isActiveParam !== null && isActiveParam !== undefined && isActiveParam !== "") where.isActive = isActiveParam === "true";

  const [advertisements, total] = await Promise.all([
    prisma.advertisement.findMany({ where, orderBy: { createdAt: "desc" }, skip, take: limit }),
    prisma.advertisement.count({ where }),
  ]);

  return apiPaginated(advertisements, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const { image, imageWidth = 0, imageHeight = 0, link, title, position, isActive = false } = body;

    if (!image || !link || !title || !position) return apiError("请填写所有必填字段", 400);

    const existingAd = await prisma.advertisement.findUnique({ where: { position } });
    if (existingAd) return apiError("该广告位已被占用", 409);

    const advertisement = await prisma.advertisement.create({
      data: { image, imageWidth, imageHeight, link, title, position, isActive },
    });

    revalidatePath("/console/advertisements");
    return apiSuccess(advertisement, 201);
  } catch (error) {
    console.error("Error creating advertisement:", error);
    return apiError("创建广告失败");
  }
}
