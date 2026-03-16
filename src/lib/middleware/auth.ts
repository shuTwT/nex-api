import { NextRequest, NextResponse } from "next/server";
import { verify } from "jsonwebtoken";

const JWT_SECRET = process.env.JWT_SECRET || "your-secret-key";

export interface AuthUser {
  id: string;
  email: string;
  username: string;
  role: string;
}

export function withAuth(
  handler: (req: NextRequest, user: AuthUser) => Promise<NextResponse>
) {
  return async (req: NextRequest) => {
    try {
      const authHeader = req.headers.get("authorization");
      
      if (!authHeader || !authHeader.startsWith("Bearer ")) {
        return NextResponse.json(
          { success: false, error: "Unauthorized" },
          { status: 401 }
        );
      }

      const token = authHeader.substring(7);
      
      const decoded = verify(token, JWT_SECRET) as AuthUser;
      
      return await handler(req, decoded);
    } catch (error) {
      console.error("Auth error:", error);
      return NextResponse.json(
        { success: false, error: "Invalid token" },
        { status: 401 }
      );
    }
  };
}

export function withAdminAuth(
  handler: (req: NextRequest, user: AuthUser) => Promise<NextResponse>
) {
  return withAuth(async (req, user) => {
    if (user.role !== "admin") {
      return NextResponse.json(
        { success: false, error: "Forbidden: Admin access required" },
        { status: 403 }
      );
    }
    
    return await handler(req, user);
  });
}

export function withRoleAuth(
  allowedRoles: string[],
  handler: (req: NextRequest, user: AuthUser) => Promise<NextResponse>
) {
  return withAuth(async (req, user) => {
    if (!allowedRoles.includes(user.role)) {
      return NextResponse.json(
        { success: false, error: "Forbidden: Insufficient permissions" },
        { status: 403 }
      );
    }
    
    return await handler(req, user);
  });
}
