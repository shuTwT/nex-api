import { NextRequest, NextResponse } from "next/server";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { getGlobalRequestCount, getAllApiRequestCounts, getApiRequestCount, getUserApiRequestCount } from "@/lib/request-stats";

export async function GET(request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { searchParams } = new URL(request.url);
    const type = searchParams.get("type");

    if (type === "global") {
      const count = await getGlobalRequestCount();
      return apiSuccess({ totalRequests: count });
    }

    if (type === "all") {
      const stats = await getAllApiRequestCounts();
      return apiSuccess(stats);
    }

    return apiSuccess({ message: "Specify type=global or type=all" });
  } catch (error) {
    console.error("Error getting stats:", error);
    return apiError("获取统计失败");
  }
}
