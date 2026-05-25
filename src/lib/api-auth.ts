import { getServerSession } from "next-auth";
import { NextResponse } from "next/server";
import { authOptions } from "@/lib/auth/config";
import type { SessionUser } from "@/lib/auth/util";

/**
 * Require authentication for API routes.
 * Returns the session user or a 401 NextResponse.
 */
export async function getAuthUser(): Promise<SessionUser | NextResponse> {
  const session = await getServerSession(authOptions);

  if (!session?.user) {
    return NextResponse.json(
      { success: false, error: "Unauthorized" },
      { status: 401 }
    );
  }

  return session.user as SessionUser;
}

/**
 * Require admin role for API routes.
 * Returns the session user or a 401/403 NextResponse.
 */
export async function getAdminUser(): Promise<SessionUser | NextResponse> {
  const session = await getServerSession(authOptions);

  if (!session?.user) {
    return NextResponse.json(
      { success: false, error: "Unauthorized" },
      { status: 401 }
    );
  }

  const user = session.user as SessionUser;

  // Admin check is currently commented out in the source.
  // Uncomment if admin enforcement is needed:
  // if (user.role !== "admin") {
  //   return NextResponse.json(
  //     { success: false, error: "Forbidden: Admin access required" },
  //     { status: 403 }
  //   );
  // }

  return user;
}

/**
 * Standard API error response.
 */
export function apiError(message: string, status = 500): NextResponse {
  return NextResponse.json({ success: false, error: message }, { status });
}

/**
 * Standard API success response.
 */
export function apiSuccess<T>(data: T, status = 200): NextResponse {
  return NextResponse.json({ success: true, data }, { status });
}

/**
 * Standard API paginated response.
 */
export function apiPaginated<T>(
  data: T[],
  pagination: { page: number; limit: number; total: number }
): NextResponse {
  return NextResponse.json({
    success: true,
    data,
    pagination: {
      ...pagination,
      totalPages: Math.ceil(pagination.total / pagination.limit),
    },
  });
}
