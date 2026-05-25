import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { getAuthUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  const sessionUser = await getAuthUser();
  if (sessionUser instanceof NextResponse) return sessionUser;

  try {
    const payments = await prisma.payment.findMany({
      where: { userId: sessionUser.id },
      orderBy: { createdAt: "desc" },
    });
    return apiSuccess(payments);
  } catch (error) {
    console.error("Error fetching user payments:", error);
    return apiError("获取支付记录失败");
  }
}
