"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/auth/util";
import { AdPosition } from "@/types/ad-position";
import { authOptions } from "@/lib/auth/config";

export interface Advertisement {
  id: string;
  image: string;
  imageWidth: number;
  imageHeight: number;
  link: string;
  title: string;
  position: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export async function getAdvertisements(params?: {
  position?: string;
  isActive?: boolean;
  page?: number;
  limit?: number;
}) {
  try {
    const { position, isActive, page = 1, limit = 10 } = params || {};
    const skip = (page - 1) * limit;

    const where: any = {};

    if (position) {
      where.position = position;
    }

    if (isActive !== undefined) {
      where.isActive = isActive;
    }

    const [advertisements, total] = await Promise.all([
      prisma.advertisement.findMany({
        where,
        orderBy: { createdAt: "desc" },
        skip,
        take: limit,
      }),
      prisma.advertisement.count({ where }),
    ]);

    return {
      success: true,
      data: advertisements,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Error fetching advertisements:", error);
    return { success: false, error: "Failed to fetch advertisements" };
  }
}

export async function getAdvertisementById(id: string) {
  try {
    const advertisement = await prisma.advertisement.findUnique({
      where: { id },
    });

    if (!advertisement) {
      return { success: false, error: "Advertisement not found" };
    }

    return { success: true, data: advertisement };
  } catch (error) {
    console.error("Error fetching advertisement:", error);
    return { success: false, error: "Failed to fetch advertisement" };
  }
}

export async function getAdvertisementByPosition(position: AdPosition) {
  try {
    const advertisement = await prisma.advertisement.findUnique({
      where: { position },
    });

    return { success: true, data: advertisement };
  } catch (error) {
    console.error("Error fetching advertisement by position:", error);
    return { success: false, error: "Failed to fetch advertisement" };
  }
}

export async function createAdvertisement(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const data = {
      image: formData.get("image") as string,
      imageWidth: parseInt(formData.get("imageWidth") as string) || 0,
      imageHeight: parseInt(formData.get("imageHeight") as string) || 0,
      link: formData.get("link") as string,
      title: formData.get("title") as string,
      position: formData.get("position") as string,
      isActive: formData.get("isActive") === "true",
    };

    if (!data.image || !data.link || !data.title || !data.position) {
      return {
        success: false,
        error: "请填写所有必填字段",
      };
    }

    const existingAd = await prisma.advertisement.findUnique({
      where: { position: data.position },
    });

    if (existingAd) {
      return {
        success: false,
        error: "该广告位已被占用",
      };
    }

    const advertisement = await prisma.advertisement.create({
      data,
    });

    revalidatePath("/console/advertisements");
    return { success: true, data: advertisement };
  } catch (error) {
    console.error("Error creating advertisement:", error);

    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "创建广告失败" };
  }
}

export async function updateAdvertisement(formData: FormData) {
  try {
    await requireAdmin(authOptions);

    const id = formData.get("id") as string;
    const position = formData.get("position") as string;

    const data: any = {
      image: formData.get("image") as string || undefined,
      imageWidth: formData.get("imageWidth") ? parseInt(formData.get("imageWidth") as string) : undefined,
      imageHeight: formData.get("imageHeight") ? parseInt(formData.get("imageHeight") as string) : undefined,
      link: formData.get("link") as string || undefined,
      title: formData.get("title") as string || undefined,
      position: position || undefined,
      isActive: formData.get("isActive") === "true",
    };

    if (position) {
      const existingAd = await prisma.advertisement.findFirst({
        where: {
          AND: [
            { id: { not: id } },
            { position },
          ],
        },
      });

      if (existingAd) {
        return {
          success: false,
          error: "该广告位已被占用",
        };
      }
    }

    const advertisement = await prisma.advertisement.update({
      where: { id },
      data,
    });

    revalidatePath("/console/advertisements");
    return { success: true, data: advertisement };
  } catch (error) {
    console.error("Error updating advertisement:", error);

    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "更新广告失败" };
  }
}

export async function deleteAdvertisement(id: string) {
  try {
    await requireAdmin(authOptions);

    await prisma.advertisement.delete({
      where: { id },
    });

    revalidatePath("/console/advertisements");
    return { success: true, message: "广告已删除" };
  } catch (error) {
    console.error("Error deleting advertisement:", error);

    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "删除广告失败" };
  }
}

export async function toggleAdvertisementStatus(id: string) {
  try {
    await requireAdmin(authOptions);

    const advertisement = await prisma.advertisement.findUnique({
      where: { id },
      select: { isActive: true },
    });

    if (!advertisement) {
      return { success: false, error: "广告不存在" };
    }

    const updated = await prisma.advertisement.update({
      where: { id },
      data: { isActive: !advertisement.isActive },
    });

    revalidatePath("/console/advertisements");
    return { success: true, data: updated };
  } catch (error) {
    console.error("Error toggling advertisement status:", error);

    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "切换广告状态失败" };
  }
}

export async function getAdvertisementStats() {
  try {
    const [totalAds, activeAds, positionStats] = await Promise.all([
      prisma.advertisement.count(),
      prisma.advertisement.count({
        where: { isActive: true },
      }),
      prisma.advertisement.groupBy({
        by: ["position"],
        _count: true,
      }),
    ]);

    return {
      success: true,
      data: {
        totalAds,
        activeAds,
        inactiveAds: totalAds - activeAds,
        positionStats,
      },
    };
  } catch (error) {
    console.error("Error fetching advertisement stats:", error);
    return { success: false, error: "获取广告统计失败" };
  }
}
