import { NextRequest } from "next/server";
import prisma from "@/lib/prisma";

export async function GET(_request: NextRequest) {
  try {
    const userCount = await prisma.user.count();
    return Response.json({ initialized: userCount > 0 });
  } catch (error) {
    console.error("Check system initialized error:", error);
    return Response.json({ initialized: false });
  }
}
