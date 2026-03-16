import { NextResponse } from "next/server";
import prisma from "@/lib/prisma";

export async function GET() {
  try {
    const apis = await prisma.api.findMany({
      include: {
        category: true,
        parameters: true,
        responses: true,
      },
    });

    return NextResponse.json({
      success: true,
      data: apis,
    });
  } catch (error) {
    console.error("Error fetching APIs:", error);
    return NextResponse.json(
      {
        success: false,
        error: "Failed to fetch APIs",
      },
      { status: 500 }
    );
  }
}
