import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { logAudit } from "@/lib/audit-log";
import { incrementRequestCount } from "@/lib/request-stats";

const MCP_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type, Authorization",
};

interface RouteParams {
  params: Promise<{ identifier: string }>;
}

async function verifyApiToken(
  request: NextRequest
): Promise<{ userId: string } | null> {
  const authHeader = request.headers.get("authorization");
  if (!authHeader) return null;

  const token = authHeader.replace(/^Bearer\s+/i, "");
  if (!token) return null;

  const apiToken = await prisma.apiToken.findUnique({
    where: { token },
    select: { id: true, userId: true, isActive: true, expiresAt: true },
  });

  if (!apiToken || !apiToken.isActive) return null;
  if (apiToken.expiresAt && new Date() > apiToken.expiresAt) return null;

  await prisma.apiToken.update({
    where: { id: apiToken.id },
    data: { lastUsedAt: new Date() },
  });

  return { userId: apiToken.userId };
}

async function deductCredits(
  userId: string,
  mcpId: string,
  credits: number
): Promise<boolean> {
  try {
    await prisma.$transaction(async (tx) => {
      const user = await tx.user.findUnique({
        where: { id: userId },
        select: { credits: true },
      });
      if (!user || user.credits < credits) throw new Error("积分不足");

      await tx.user.update({
        where: { id: userId },
        data: { credits: { decrement: credits } },
      });

      await tx.mcpUsage.create({
        data: { userId, mcpId, credits, status: "success" },
      });

      await tx.mcpService.update({
        where: { id: mcpId },
        data: { callCount: { increment: 1 } },
      });
    });
    return true;
  } catch (error) {
    console.error("MCP deduct credits error:", error);
    return false;
  }
}

export async function OPTIONS() {
  return new NextResponse(null, { status: 204, headers: MCP_HEADERS });
}

export async function POST(
  request: NextRequest,
  { params }: RouteParams
) {
  const ipAddress = request.headers.get("x-forwarded-for") || "unknown";
  const userAgent = request.headers.get("user-agent") || "unknown";

  try {
    const { identifier } = await params;

    const tokenInfo = await verifyApiToken(request);
    if (!tokenInfo) {
      await logAudit({
        action: "MCP 调用失败",
        resource: `MCP: ${identifier}`,
        details: "无效的 API Token",
        ipAddress,
        userAgent,
        level: "warning",
        status: "error",
      });
      return NextResponse.json(
        { success: false, error: "无效的 API Token" },
        { status: 401, headers: MCP_HEADERS }
      );
    }

    const mcpService = await prisma.mcpService.findUnique({
      where: { identifier },
      select: {
        id: true,
        name: true,
        type: true,
        command: true,
        endpoint: true,
        envVars: true,
        pricing: true,
        isActive: true,
      },
    });

    if (!mcpService) {
      await logAudit({
        userId: tokenInfo.userId,
        action: "MCP 调用失败",
        resource: `MCP: ${identifier}`,
        details: "MCP 服务不存在",
        ipAddress,
        userAgent,
        level: "warning",
        status: "error",
      });
      return NextResponse.json(
        { success: false, error: "MCP 服务不存在" },
        { status: 404, headers: MCP_HEADERS }
      );
    }

    if (!mcpService.isActive) {
      return NextResponse.json(
        { success: false, error: "MCP 服务已禁用" },
        { status: 403, headers: MCP_HEADERS }
      );
    }

    const user = await prisma.user.findUnique({
      where: { id: tokenInfo.userId },
      select: { credits: true },
    });

    if (!user || user.credits < mcpService.pricing) {
      return NextResponse.json(
        { success: false, error: "积分不足" },
        { status: 402, headers: MCP_HEADERS }
      );
    }

    let clientRequestBody: unknown = null;
    try {
      clientRequestBody = await request.json();
    } catch {
    }

    const env = mcpService.envVars ? JSON.parse(String(mcpService.envVars)) : {};

    let stream: ReadableStream<Uint8Array>;

    switch (mcpService.type) {
      case "streamableHttp": {
        if (!mcpService.endpoint) {
          return NextResponse.json(
            { success: false, error: "streamableHttp 类型缺少端点配置" },
            { status: 500, headers: MCP_HEADERS }
          );
        }
        const upstreamResponse = await fetch(mcpService.endpoint, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(env as Record<string, string>),
          },
          body: clientRequestBody ? JSON.stringify(clientRequestBody) : undefined,
        });
        if (!upstreamResponse.ok || !upstreamResponse.body) {
          return NextResponse.json(
            { success: false, error: `上游服务返回错误: ${upstreamResponse.status}` },
            { status: 502, headers: MCP_HEADERS }
          );
        }
        stream = upstreamResponse.body;
        break;
      }

      case "sse": {
        if (!mcpService.endpoint) {
          return NextResponse.json(
            { success: false, error: "SSE 类型缺少端点配置" },
            { status: 500, headers: MCP_HEADERS }
          );
        }
        const sseResponse = await fetch(mcpService.endpoint, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Accept: "text/event-stream",
            ...(env as Record<string, string>),
          },
          body: clientRequestBody ? JSON.stringify(clientRequestBody) : undefined,
        });

        if (!sseResponse.ok || !sseResponse.body) {
          return NextResponse.json(
            { success: false, error: `SSE 上游服务返回错误: ${sseResponse.status}` },
            { status: 502, headers: MCP_HEADERS }
          );
        }

        stream = sseToJsonStream(sseResponse.body);
        break;
      }

      case "stdio": {
        if (!mcpService.command) {
          return NextResponse.json(
            { success: false, error: "stdio 类型缺少启动命令配置" },
            { status: 500, headers: MCP_HEADERS }
          );
        }
        stream = stdioToJsonStream(
          mcpService.command,
          env as Record<string, string>,
          clientRequestBody
        );
        break;
      }

      default:
        return NextResponse.json(
          { success: false, error: `不支持的服务类型: ${mcpService.type}` },
          { status: 400, headers: MCP_HEADERS }
        );
    }

    const deducted = await deductCredits(
      tokenInfo.userId,
      mcpService.id,
      mcpService.pricing
    );
    if (!deducted) {
      await logAudit({
        userId: tokenInfo.userId,
        action: "MCP 调用失败",
        resource: mcpService.name,
        details: "积分扣除失败",
        ipAddress,
        userAgent,
        level: "error",
        status: "error",
      });
      return NextResponse.json(
        { success: false, error: "积分扣除失败" },
        { status: 500, headers: MCP_HEADERS }
      );
    }

    await incrementRequestCount(
      tokenInfo.userId,
      `mcp:${identifier}`,
      mcpService.pricing
    );

    await logAudit({
      userId: tokenInfo.userId,
      action: "MCP 调用",
      resource: mcpService.name,
      details: `成功调用 MCP 服务，类型: ${mcpService.type}`,
      ipAddress,
      userAgent,
      level: "info",
      status: "success",
    });

    return new NextResponse(stream, {
      status: 200,
      headers: {
        ...MCP_HEADERS,
        "Content-Type": "application/json; charset=utf-8",
      },
    });
  } catch (error) {
    console.error("MCP gateway error:", error);
    await logAudit({
      action: "MCP 调用失败",
      resource: "MCP 网关",
      details: `服务器内部错误: ${error}`,
      ipAddress,
      userAgent,
      level: "error",
      status: "error",
    });
    return NextResponse.json(
      { success: false, error: "服务器内部错误" },
      { status: 500, headers: MCP_HEADERS }
    );
  }
}

function sseToJsonStream(
  sseBody: ReadableStream<Uint8Array>
): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  let buffer = "";

  return new ReadableStream({
    async start(controller) {
      const reader = sseBody.getReader();
      const decoder = new TextDecoder();

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            if (buffer.trim()) {
              controller.enqueue(encodeSSELine(buffer.trim()));
            }
            controller.close();
            return;
          }

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed.startsWith("data:")) {
              const data = trimmed.slice(5).trim();
              if (data && data !== "[DONE]") {
                controller.enqueue(encoder.encode(data + "\n"));
              }
            }
          }
        }
      } catch (error) {
        controller.error(error);
      }
    },
  });
}

function stdioToJsonStream(
  command: string,
  env: Record<string, string>,
  requestBody: unknown
): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  let spawned: ReturnType<typeof import("child_process").spawn> | undefined;

  return new ReadableStream({
    async start(controller) {
      try {
        const { spawn } = await import("node:child_process");
        const args = command.split(/\s+/).filter(Boolean);
        const cmd = args[0];
        const cmdArgs = args.slice(1);

        const mergedEnv = { ...process.env, ...env };

        spawned = spawn(cmd!, cmdArgs, {
          env: mergedEnv,
          stdio: ["pipe", "pipe", "pipe"],
          shell: true,
        });

        if (spawned.stdout) {
          spawned.stdout.on("data", (chunk: Buffer) => {
            controller.enqueue(chunk);
          });

          spawned.stdout.on("end", () => {
            controller.close();
          });

          spawned.stdout.on("error", (err: Error) => {
            controller.error(err);
          });
        }

        if (spawned.stderr) {
          spawned.stderr.on("data", (chunk: Buffer) => {
            console.error("MCP stdio stderr:", chunk.toString());
          });
        }

        spawned.on("error", (err: Error) => {
          console.error("MCP stdio spawn error:", err);
          controller.error(err);
        });

        spawned.on("exit", (code: number | null) => {
          if (code !== 0 && code !== null) {
            console.error(`MCP stdio process exited with code ${code}`);
          }
          if (spawned?.stdout && !spawned.stdout.destroyed) {
            controller.close();
          }
        });

        if (requestBody && spawned.stdin) {
          spawned.stdin.write(JSON.stringify(requestBody) + "\n");
          spawned.stdin.end();
        }
      } catch (error) {
        controller.error(error);
      }
    },

    cancel() {
      if (spawned && !spawned.killed) {
        spawned.kill();
      }
    },
  });
}

function encodeSSELine(data: string): Uint8Array {
  const encoder = new TextEncoder();
  return encoder.encode(data + "\n");
}
