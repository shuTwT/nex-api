import { NextResponse } from "next/server";
import prisma from "@/lib/prisma";

export async function GET() {
  try {
    const [totalUsers, activeUsers, adminUsers, newUsersThisMonth] = await Promise.all([
      prisma.user.count(),
      prisma.user.count({
        where: {
          apiUsage: {
            some: {
              createdAt: {
                gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
              },
            },
          },
        },
      }),
      prisma.user.count({
        where: { role: "admin" },
      }),
      prisma.user.count({
        where: {
          createdAt: {
            gte: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
          },
        },
      }),
    ]);

    return NextResponse.json({
      success: true,
      data: {
        totalUsers,
        activeUsers,
        adminUsers,
        newUsersThisMonth,
      },
    });
  } catch (error) {
    console.error("Error fetching user stats:", error);
    return NextResponse.json(
      { success: false, error: "Failed to fetch user stats" },
      { status: 500 }
    );
  }
}
