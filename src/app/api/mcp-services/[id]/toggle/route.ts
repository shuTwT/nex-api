import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";

export async function PUT(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const service = await prisma.mcpService.findUnique({ where: { id } });
    if (!service) return apiError("MCP 服务不存在", 404);

    const updated = await prisma.mcpService.update({
      where: { id },
      data: { isActive: !service.isActive },
    });

    revalidatePath("/console/mcp-services");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Toggle MCP service status error:", error);
    return apiError("切换 MCP 服务状态失败");
  }
}
