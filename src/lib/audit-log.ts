
import prisma from "@/lib/prisma";

interface LogAuditParams {
  userId?: string;
  action: string;
  resource: string;
  details?: string;
  ipAddress?: string;
  userAgent?: string;
  level?: "info" | "warning" | "error";
  status?: "success" | "error";
  metadata?: string;
}

export async function logAudit({
  userId,
  action,
  resource,
  details,
  ipAddress,
  userAgent,
  level = "info",
  status = "success",
  metadata,
}: LogAuditParams) {
  try {
    await prisma.auditLog.create({
      data: {
        userId,
        action,
        resource,
        details,
        ipAddress,
        userAgent,
        level,
        status,
        metadata,
      },
    });
  } catch (error) {
    console.error("Failed to log audit:", error);
  }
}
