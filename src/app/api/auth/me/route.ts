import { NextResponse } from "next/server";
import { getCurrentUser } from "@/lib/session";
import prisma from "@/lib/prisma";

export async function GET() {
  try {
    const sessionUser = await getCurrentUser();

    if (!sessionUser) {
      return NextResponse.json(
        { success: false, error: "Not authenticated" },
        { status: 401 }
      );
    }

    const user = await prisma.user.findUnique({
      where: { id: sessionUser.id },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
      },
    });

    if (!user) {
      return NextResponse.json(
        { success: false, error: "User not found" },
        { status: 404 }
      );
    }

    return NextResponse.json({
      success: true,
      data: user,
    });
  } catch (error) {
    console.error("Get current user error:", error);
    return NextResponse.json(
      { success: false, error: "Failed to get current user" },
      { status: 500 }
    );
  }
}
