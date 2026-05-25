import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const body = await request.json();
    const {
      name,
      identifier,
      type,
      command,
      endpoint,
      envVars,
      pricing,
      isActive,
    } = body;

    if (!name || !identifier || !type) {
      return apiError("缺少必填字段", 400);
    }

    const validTypes = ["stdio", "sse", "streamableHttp"];
    if (!validTypes.includes(type)) {
      return apiError("无效的服务类型", 400);
    }

    const identifierPattern = /^[a-zA-Z][a-zA-Z0-9-]*$/;
    if (!identifierPattern.test(identifier)) {
      return apiError("标识必须以字母开头，只能包含字母、数字和连字符", 400);
    }

    const existingIdentifier = await prisma.mcpService.findFirst({
      where: { identifier, NOT: { id } },
    });
    if (existingIdentifier) return apiError("标识已存在", 409);

    const updated = await prisma.mcpService.update({
      where: { id },
      data: {
        name,
        identifier,
        type,
        command: type === "stdio" ? command : null,
        endpoint: type !== "stdio" ? endpoint : null,
        envVars,
        pricing: parseInt(String(pricing)) || 0,
        isActive,
      },
    });

    await logAudit({
      action: "更新 MCP 服务",
      resource: "MCP 服务管理",
      details: `更新了 MCP 服务: ${name} (${identifier})`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/mcp-services");
    return apiSuccess(updated);
  } catch (error) {
    console.error("Update MCP service error:", error);
    await logAudit({
      action: "更新 MCP 服务",
      resource: "MCP 服务管理",
      details: `更新 MCP 服务失败: ${error}`,
      level: "error",
      status: "error",
    });
    return apiError("更新 MCP 服务失败");
  }
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const { id } = await params;
    const service = await prisma.mcpService.findUnique({ where: { id } });
    await prisma.mcpService.delete({ where: { id } });

    await logAudit({
      action: "删除 MCP 服务",
      resource: "MCP 服务管理",
      details: `删除了 MCP 服务: ${service?.name || id}`,
      level: "warning",
      status: "success",
    });

    revalidatePath("/console/mcp-services");
    return apiSuccess({ message: "MCP 服务已删除" });
  } catch (error) {
    console.error("Delete MCP service error:", error);
    await logAudit({
      action: "删除 MCP 服务",
      resource: "MCP 服务管理",
      details: `删除 MCP 服务失败: ${error}`,
      level: "error",
      status: "error",
    });
    return apiError("删除 MCP 服务失败");
  }
}
