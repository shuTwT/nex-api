import { NextRequest, NextResponse } from "next/server";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";
import { getApiRequestCount, getUserApiRequestCount } from "@/lib/request-stats";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ alias: string }> }
) {
  const user = await getAuthUser();
  if (user instanceof NextResponse) return user;

  try {
    const { alias } = await params;
    const { searchParams } = new URL(request.url);

    if (searchParams.get("user") === "true") {
      const count = await getUserApiRequestCount(user.id, alias);
      return apiSuccess({ apiAlias: alias, requestCount: count });
    }

    const count = await getApiRequestCount(alias);
    return apiSuccess({ apiAlias: alias, requestCount: count });
  } catch (error) {
    console.error("Error getting API stats:", error);
    return apiError("获取 API 统计失败");
  }
}
