import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { executePreScript, executePostScript } from "@/lib/sandbox";

interface RouteParams {
  params: Promise<{ alias: string }>;
}

async function verifyApiToken(request: NextRequest): Promise<{ userId: string } | null> {
  const authHeader = request.headers.get("authorization");
  
  if (!authHeader) {
    return null;
  }

  const token = authHeader.replace(/^Bearer\s+/i, "");
  
  if (!token) {
    return null;
  }

  const apiToken = await prisma.apiToken.findUnique({
    where: { token },
    select: {
      id: true,
      userId: true,
      isActive: true,
      expiresAt: true,
    },
  });

  if (!apiToken || !apiToken.isActive) {
    return null;
  }

  if (apiToken.expiresAt && new Date() > apiToken.expiresAt) {
    return null;
  }

  await prisma.apiToken.update({
    where: { id: apiToken.id },
    data: { lastUsedAt: new Date() },
  });

  return { userId: apiToken.userId };
}

async function deductCredits(userId: string, apiId: string, credits: number): Promise<boolean> {
  try {
    await prisma.$transaction(async (tx) => {
      const user = await tx.user.findUnique({
        where: { id: userId },
        select: { credits: true },
      });

      if (!user || user.credits < credits) {
        throw new Error("积分不足");
      }

      await tx.user.update({
        where: { id: userId },
        data: { credits: { decrement: credits } },
      });

      await tx.apiUsage.create({
        data: {
          userId,
          apiId,
          credits,
          status: "success",
        },
      });

      await tx.api.update({
        where: { id: apiId },
        data: { callCount: { increment: 1 } },
      });
    });

    return true;
  } catch (error) {
    console.error("Deduct credits error:", error);
    return false;
  }
}

export async function GET(request: NextRequest, { params }: RouteParams) {
  return handleRequest(request, params);
}

export async function POST(request: NextRequest, { params }: RouteParams) {
  return handleRequest(request, params);
}

export async function PUT(request: NextRequest, { params }: RouteParams) {
  return handleRequest(request, params);
}

export async function DELETE(request: NextRequest, { params }: RouteParams) {
  return handleRequest(request, params);
}

export async function PATCH(request: NextRequest, { params }: RouteParams) {
  return handleRequest(request, params);
}

async function handleRequest(request: NextRequest, params: Promise<{ alias: string }>) {
  try {
    const { alias } = await params;

    const tokenInfo = await verifyApiToken(request);
    if (!tokenInfo) {
      return NextResponse.json(
        { success: false, error: "无效的 API Token" },
        { status: 401 }
      );
    }

    const api = await prisma.api.findUnique({
      where: { alias },
      select: {
        id: true,
        name: true,
        endpoint: true,
        method: true,
        pricing: true,
        preScript: true,
        postScript: true,
        isActive: true,
      },
    });

    if (!api) {
      return NextResponse.json(
        { success: false, error: "API 不存在" },
        { status: 404 }
      );
    }

    if (!api.isActive) {
      return NextResponse.json(
        { success: false, error: "API 已禁用" },
        { status: 403 }
      );
    }

    const requestMethod = request.method.toUpperCase();
    if (api.method.toUpperCase() !== requestMethod && api.method.toUpperCase() !== "ALL") {
      return NextResponse.json(
        { success: false, error: `不支持的请求方法: ${requestMethod}` },
        { status: 405 }
      );
    }

    const headers: Record<string, string> = {};
    request.headers.forEach((value, key) => {
      if (!["host", "authorization", "content-length"].includes(key.toLowerCase())) {
        headers[key] = value;
      }
    });

    const query: Record<string, string> = {};
    request.nextUrl.searchParams.forEach((value, key) => {
      query[key] = value;
    });

    let body: unknown = null;
    if (["POST", "PUT", "PATCH"].includes(requestMethod)) {
      try {
        const contentType = request.headers.get("content-type");
        if (contentType?.includes("application/json")) {
          body = await request.json();
        } else if (contentType?.includes("multipart/form-data")) {
          body = await request.formData();
        } else {
          body = await request.text();
        }
      } catch {
        body = null;
      }
    }

    let finalHeaders = headers;
    let finalQuery = query;
    let finalBody = body;

    if (api.preScript) {
      try {
        const preScriptResult = executePreScript(api.preScript, {
          headers,
          query,
          body,
        });

        if (preScriptResult.headers) {
          finalHeaders = { ...headers, ...preScriptResult.headers };
        }
        if (preScriptResult.query) {
          finalQuery = { ...query, ...preScriptResult.query };
        }
        if (preScriptResult.body !== undefined) {
          finalBody = preScriptResult.body;
        }
      } catch (error) {
        console.error("PreScript error:", error);
        return NextResponse.json(
          { success: false, error: error instanceof Error ? error.message : "预处理脚本执行失败" },
          { status: 500 }
        );
      }
    }

    const user = await prisma.user.findUnique({
      where: { id: tokenInfo.userId },
      select: { credits: true },
    });

    if (!user || user.credits < api.pricing) {
      return NextResponse.json(
        { success: false, error: "积分不足" },
        { status: 402 }
      );
    }

    const upstreamUrl = new URL(api.endpoint);
    Object.entries(finalQuery).forEach(([key, value]) => {
      upstreamUrl.searchParams.set(key, value);
    });

    const upstreamHeaders = new Headers();
    Object.entries(finalHeaders).forEach(([key, value]) => {
      upstreamHeaders.set(key, value);
    });

    const upstreamRequest: RequestInit = {
      method: requestMethod,
      headers: upstreamHeaders,
    };

    if (["POST", "PUT", "PATCH"].includes(requestMethod) && finalBody) {
      if (typeof finalBody === "string") {
        upstreamRequest.body = finalBody;
      } else {
        upstreamRequest.body = JSON.stringify(finalBody);
        if (!upstreamHeaders.has("content-type")) {
          upstreamHeaders.set("content-type", "application/json");
        }
      }
    }

    console.log("准备发送请求", upstreamUrl.toString(), upstreamRequest);
    const upstreamResponse = await fetch(upstreamUrl.toString(), upstreamRequest);

    let responseBody: unknown;
    const responseHeaders: Record<string, string> = {};
    
    upstreamResponse.headers.forEach((value, key) => {
      responseHeaders[key] = value;
    });

    const responseContentType = upstreamResponse.headers.get("content-type");
    if (responseContentType?.includes("application/json")) {
      try {
        responseBody = await upstreamResponse.json();
      } catch {
        responseBody = await upstreamResponse.text();
      }
    } else {
      responseBody = await upstreamResponse.text();
    }

    console.log("收到响应", upstreamResponse.status, upstreamResponse.statusText, responseBody);

    let finalResponseBody = responseBody;
    let finalResponseHeaders = responseHeaders;

    if (api.postScript) {
      try {
        const postScriptResult = executePostScript(api.postScript, {
          responseBody,
          responseHeaders,
        });

        if (postScriptResult.responseBody !== undefined) {
          finalResponseBody = postScriptResult.responseBody;
        }
        if (postScriptResult.responseHeaders) {
          finalResponseHeaders = { ...responseHeaders, ...postScriptResult.responseHeaders };
        }
      } catch (error) {
        console.error("PostScript error:", error);
        return NextResponse.json(
          { success: false, error: error instanceof Error ? error.message : "后处理脚本执行失败" },
          { status: 500 }
        );
      }
    }

    const deducted = await deductCredits(tokenInfo.userId, api.id, api.pricing);
    if (!deducted) {
      return NextResponse.json(
        { success: false, error: "积分扣除失败" },
        { status: 500 }
      );
    }

    const response = NextResponse.json(finalResponseBody, {
      status: upstreamResponse.status,
    });

    Object.entries(finalResponseHeaders).forEach(([key, value]) => {
      if (!["content-encoding", "transfer-encoding", "connection"].includes(key.toLowerCase())) {
        response.headers.set(key, value);
      }
    });

    return response;
  } catch (error) {
    console.error("API proxy error:", error);
    return NextResponse.json(
      { success: false, error: "服务器内部错误" },
      { status: 500 }
    );
  }
}
