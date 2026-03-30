import { NextRequest, NextResponse } from "next/server";
import { syncToDatabase } from "@/lib/request-stats";

export async function POST(request: NextRequest) {
  try {
    const authHeader = request.headers.get("authorization");
    const cronSecret = process.env.CRON_SECRET;
    
    if (cronSecret && authHeader !== `Bearer ${cronSecret}`) {
      return NextResponse.json(
        { success: false, error: "未授权" },
        { status: 401 }
      );
    }

    await syncToDatabase();

    return NextResponse.json({
      success: true,
      message: "数据同步成功",
    });
  } catch (error) {
    console.error("Sync error:", error);
    return NextResponse.json(
      { success: false, error: "数据同步失败" },
      { status: 500 }
    );
  }
}
