import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const advertisement = await prisma.advertisement.findUnique({ where: { id } });
  if (!advertisement) return apiError("Advertisement not found", 404);
  return apiSuccess(advertisement);
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();
    const { image, imageWidth, imageHeight, link, title, position, isActive } = body;

    if (position) {
      const existingAd = await prisma.advertisement.findFirst({
        where: { AND: [{ id: { not: id } }, { position }] },
      });
      if (existingAd) return apiError("该广告位已被占用", 409);
    }

    const data: Record<string, unknown> = {};
    if (image !== undefined) data.image = image;
    if (imageWidth !== undefined) data.imageWidth = imageWidth;
    if (imageHeight !== undefined) data.imageHeight = imageHeight;
    if (link !== undefined) data.link = link;
    if (title !== undefined) data.title = title;
    if (position !== undefined) data.position = position;
    if (isActive !== undefined) data.isActive = isActive;

    const advertisement = await prisma.advertisement.update({ where: { id }, data });

    revalidatePath("/console/advertisements");
    return apiSuccess(advertisement);
  } catch (error) {
    console.error("Error updating advertisement:", error);
    return apiError("更新广告失败");
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    await prisma.advertisement.delete({ where: { id } });

    revalidatePath("/console/advertisements");
    return apiSuccess({ message: "广告已删除" });
  } catch (error) {
    console.error("Error deleting advertisement:", error);
    return apiError("删除广告失败");
  }
}
