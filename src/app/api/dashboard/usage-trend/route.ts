import { NextRequest, NextResponse } from "next/server";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { getHourlyUsageTrend } from "@/lib/request-stats";

export async function GET(_request: NextRequest) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const isAdmin = user.role === "admin";
    const now = new Date();
    const labels: string[] = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date(now);
      d.setHours(d.getHours() - i);
      labels.push(`${d.getHours()}:00`);
    }

    const userTrend = await getHourlyUsageTrend(user.id, 7);

    if (isAdmin) {
      const globalTrend = await getHourlyUsageTrend(undefined, 7);
      return apiSuccess({
        labels,
        datasets: [
          { label: "全局用量", data: globalTrend, borderColor: "rgb(59, 130, 246)", backgroundColor: "rgba(59, 130, 246, 0.1)" },
          { label: "我的用量", data: userTrend, borderColor: "rgb(16, 185, 129)", backgroundColor: "rgba(16, 185, 129, 0.1)" },
        ],
      });
    }

    return apiSuccess({
      labels,
      datasets: [
        { label: "我的用量", data: userTrend, borderColor: "rgb(59, 130, 246)", backgroundColor: "rgba(59, 130, 246, 0.1)" },
      ],
    });
  } catch (error) {
    console.error("Error getting usage trend:", error);
    return apiError("获取用量趋势失败");
  }
}
