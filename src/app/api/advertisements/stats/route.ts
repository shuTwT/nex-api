import { NextRequest } from "next/server";
import prisma from "@/lib/prisma";
import { apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  try {
    const [totalAds, activeAds, positionStats] = await Promise.all([
      prisma.advertisement.count(),
      prisma.advertisement.count({ where: { isActive: true } }),
      prisma.advertisement.groupBy({ by: ["position"], _count: true }),
    ]);

    return apiSuccess({ totalAds, activeAds, inactiveAds: totalAds - activeAds, positionStats });
  } catch (error) {
    console.error("Error fetching advertisement stats:", error);
    return apiError("获取广告统计失败");
  }
}
