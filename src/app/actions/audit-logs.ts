"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { requireAuth } from "@/lib/session";

export async function getAuditLogs(params?: {
  search?: string;
  level?: string;
  status?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  limit?: number;
}): Promise<{
  success: boolean;
  data?: any[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  error?: string;
}> {
  try {
    const user = await requireAuth();
    const { search, level, status, startDate, endDate, page = 1, limit = 10 } = params || {};
    const skip = (page - 1) * limit;

    const where: any = {};

    if (search) {
      where.OR = [
        { action: { contains: search } },
        { resource: { contains: search } },
        { details: { contains: search } },
      ];
    }

    if (level && level !== "all") {
      where.level = level;
    }

    if (status && status !== "all") {
      where.status = status;
    }

    if (startDate || endDate) {
      where.createdAt = {};
      if (startDate) {
        where.createdAt.gte = new Date(startDate);
      }
      if (endDate) {
        where.createdAt.lte = new Date(endDate);
      }
    }

    const [logs, total] = await Promise.all([
      prisma.auditLog.findMany({
        where,
        skip,
        take: limit,
        orderBy: {
          createdAt: "desc",
        },
        include: {
          user: {
            select: {
              id: true,
              name: true,
              email: true,
            },
          },
        },
      }),
      prisma.auditLog.count({ where }),
    ]);

    return {
      success: true,
      data: logs,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Get audit logs error:", error);
    return {
      success: false,
      error: "获取审计日志失败",
    };
  }
}

export async function createAuditLog(formData: FormData): Promise<{
  success: boolean;
  data?: any;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const action = formData.get("action") as string;
    const resource = formData.get("resource") as string;
    const details = formData.get("details") as string;
    const ipAddress = formData.get("ipAddress") as string;
    const userAgent = formData.get("userAgent") as string;
    const level = formData.get("level") as string;
    const status = formData.get("status") as string;
    const metadata = formData.get("metadata") as string;

    if (!action || !resource) {
      return {
        success: false,
        error: "操作和资源不能为空",
      };
    }

    const newLog = await prisma.auditLog.create({
      data: {
        userId: user.id,
        action,
        resource,
        details: details || null,
        ipAddress: ipAddress || null,
        userAgent: userAgent || null,
        level: level || "info",
        status: status || "success",
        metadata: metadata || null,
      },
    });

    revalidatePath("/console/audit-logs");
    return {
      success: true,
      data: newLog,
    };
  } catch (error) {
    console.error("Create audit log error:", error);
    return {
      success: false,
      error: "创建审计日志失败",
    };
  }
}

export async function updateAuditLog(formData: FormData): Promise<{
  success: boolean;
  data?: any;
  error?: string;
}> {
  try {
    const user = await requireAuth();

    const id = formData.get("id") as string;
    const action = formData.get("action") as string;
    const resource = formData.get("resource") as string;
    const details = formData.get("details") as string;
    const ipAddress = formData.get("ipAddress") as string;
    const userAgent = formData.get("userAgent") as string;
    const level = formData.get("level") as string;
    const status = formData.get("status") as string;
    const metadata = formData.get("metadata") as string;

    if (!id || !action || !resource) {
      return {
        success: false,
        error: "缺少必填字段",
      };
    }

    const updatedLog = await prisma.auditLog.update({
      where: { id },
      data: {
        action,
        resource,
        details: details || null,
        ipAddress: ipAddress || null,
        userAgent: userAgent || null,
        level: level || "info",
        status: status || "success",
        metadata: metadata || null,
      },
    });

    revalidatePath("/console/audit-logs");
    return {
      success: true,
      data: updatedLog,
    };
  } catch (error) {
    console.error("Update audit log error:", error);
    return {
      success: false,
      error: "更新审计日志失败",
    };
  }
}

export async function deleteAuditLog(id: string): Promise<{
  success: boolean;
  error?: string;
}> {
  try {
    await requireAuth();

    await prisma.auditLog.delete({
      where: { id },
    });

    revalidatePath("/console/audit-logs");
    return {
      success: true,
    };
  } catch (error) {
    console.error("Delete audit log error:", error);
    return {
      success: false,
      error: "删除审计日志失败",
    };
  }
}

export async function getAuditLogStats(): Promise<{
  success: boolean;
  data?: {
    totalLogs: number;
    infoLogs: number;
    warningLogs: number;
    errorLogs: number;
    successLogs: number;
    failedLogs: number;
  };
  error?: string;
}> {
  try {
    await requireAuth();

    const [totalLogs, infoLogs, warningLogs, errorLogs, successLogs, failedLogs] = await Promise.all([
      prisma.auditLog.count(),
      prisma.auditLog.count({ where: { level: "info" } }),
      prisma.auditLog.count({ where: { level: "warning" } }),
      prisma.auditLog.count({ where: { level: "error" } }),
      prisma.auditLog.count({ where: { status: "success" } }),
      prisma.auditLog.count({ where: { status: "error" } }),
    ]);

    return {
      success: true,
      data: {
        totalLogs,
        infoLogs,
        warningLogs,
        errorLogs,
        successLogs,
        failedLogs,
      },
    };
  } catch (error) {
    console.error("Get audit log stats error:", error);
    return {
      success: false,
      error: "获取审计日志统计失败",
    };
  }
}

export async function exportAuditLogs(params?: {
  level?: string;
  status?: string;
  startDate?: string;
  endDate?: string;
}): Promise<{
  success: boolean;
  data?: string;
  error?: string;
}> {
  try {
    await requireAuth();

    const { level, status, startDate, endDate } = params || {};

    const where: any = {};

    if (level && level !== "all") {
      where.level = level;
    }

    if (status && status !== "all") {
      where.status = status;
    }

    if (startDate || endDate) {
      where.createdAt = {};
      if (startDate) {
        where.createdAt.gte = new Date(startDate);
      }
      if (endDate) {
        where.createdAt.lte = new Date(endDate);
      }
    }

    const logs = await prisma.auditLog.findMany({
      where,
      orderBy: {
        createdAt: "desc",
      },
      include: {
        user: {
          select: {
            id: true,
            name: true,
            email: true,
          },
        },
      },
    });

    const csvHeader = "时间,用户,操作,资源,详情,IP地址,级别,状态\n";
    const csvRows = logs.map(log => {
      const user = log.user ? log.user.email || log.user.name || "系统" : "系统";
      return [
        log.createdAt.toISOString(),
        user,
        log.action,
        log.resource,
        log.details || "",
        log.ipAddress || "",
        log.level,
        log.status,
      ].join(",");
    }).join("\n");

    return {
      success: true,
      data: csvHeader + csvRows,
    };
  } catch (error) {
    console.error("Export audit logs error:", error);
    return {
      success: false,
      error: "导出审计日志失败",
    };
  }
}
