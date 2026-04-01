import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { AdPosition } from "@/types/ad-position";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ position: string }> }
) {
  try {
    const { position } = await params;

    const advertisement = await prisma.advertisement.findUnique({
      where: { position: position as AdPosition },
    });

    if (!advertisement) {
      return NextResponse.json({
        success: true,
        data: null,
      });
    }

    if (!advertisement.isActive) {
      return NextResponse.json({
        success: true,
        data: null,
      });
    }

    return NextResponse.json({
      success: true,
      data: advertisement,
    });
  } catch (error) {
    console.error("Error fetching advertisement:", error);
    return NextResponse.json(
      { success: false, error: "Failed to fetch advertisement" },
      { status: 500 }
    );
  }
}
