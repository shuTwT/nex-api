import { NextResponse } from "next/server";
import { clearSessionCookie } from "@/lib/session";

export async function POST() {
  try {
    await clearSessionCookie();

    return NextResponse.json({
      success: true,
      message: "Logged out successfully",
    });
  } catch (error) {
    console.error("Logout error:", error);
    return NextResponse.json(
      { success: false, error: "Logout failed" },
      { status: 500 }
    );
  }
}
