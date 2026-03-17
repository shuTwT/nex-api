import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;
    const user = await prisma.user.findUnique({
      where: { id },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
        updatedAt: true,
        subscriptions: {
          where: { isActive: true },
          select: {
            id: true,
            plan: true,
            credits: true,
            price: true,
            startDate: true,
            endDate: true,
            isActive: true,
          },
          orderBy: { createdAt: "desc" },
          take: 1,
        },
        apiUsage: {
          select: {
            id: true,
            credits: true,
            status: true,
            createdAt: true,
            api: {
              select: {
                name: true,
                endpoint: true,
              },
            },
          },
          orderBy: { createdAt: "desc" },
          take: 10,
        },
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
      data: {
        ...user,
        password: undefined,
        subscription: user.subscriptions[0] || null,
        subscriptions: undefined,
      },
    });
  } catch (error) {
    console.error("Error fetching user:", error);
    return NextResponse.json(
      { success: false, error: "Failed to fetch user" },
      { status: 500 }
    );
  }
}
