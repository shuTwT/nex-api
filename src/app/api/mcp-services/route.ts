import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";
import { logAudit } from "@/lib/audit-log";

export async function GET(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  const { searchParams } = new URL(request.url);
  const type = searchParams.get("type") || undefined;
  const search = searchParams.get("search") || undefined;
  const status = searchParams.get("status") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (type && type !== "all") where.type = type;
  if (search) {
    where.OR = [
      { name: { contains: search } },
      { identifier: { contains: search } },
    ];
  }
  if (status === "active") where.isActive = true;
  else if (status === "inactive") where.isActive = false;

  const [services, total] = await Promise.all([
    prisma.mcpService.findMany({
      where,
      skip,
      take: limit,
      orderBy: { createdAt: "desc" },
    }),
    prisma.mcpService.count({ where }),
  ]);

  return apiPaginated(services, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const {
      name,
      identifier,
      type,
      command,
      endpoint,
      envVars,
      pricing = 0,
      isActive = true,
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

    const existing = await prisma.mcpService.findUnique({
      where: { identifier },
    });
    if (existing) return apiError("标识已存在", 409);

    const service = await prisma.mcpService.create({
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
      action: "创建 MCP 服务",
      resource: "MCP 服务管理",
      details: `创建了 MCP 服务: ${name} (${identifier})`,
      level: "info",
      status: "success",
    });

    revalidatePath("/console/mcp-services");
    return apiSuccess(service, 201);
  } catch (error) {
    console.error("Create MCP service error:", error);
    await logAudit({
      action: "创建 MCP 服务",
      resource: "MCP 服务管理",
      details: `创建 MCP 服务失败: ${error}`,
      level: "error",
      status: "error",
    });
    return apiError("创建 MCP 服务失败");
  }
}
