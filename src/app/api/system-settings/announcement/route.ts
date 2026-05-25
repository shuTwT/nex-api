import { NextRequest } from "next/server";
import { apiSuccess, apiError } from "@/lib/api-auth";
import { getConfigByCategory } from "@/lib/config";

export async function GET(_request: NextRequest) {
  try {
    const operationConfig = await getConfigByCategory("operation");
    const enabled = operationConfig.announcementEnabled === "true";
    const content = operationConfig.announcementContent || "";
    return apiSuccess({ enabled, content });
  } catch (error) {
    console.error("Get public announcement error:", error);
    return apiError("获取公告失败");
  }
}
