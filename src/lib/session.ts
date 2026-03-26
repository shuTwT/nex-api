import { sign, verify } from "jsonwebtoken";
import { cookies } from "next/headers";
import { AuthUser } from "./middleware/auth";
import { getServerSession } from "next-auth";
import { authOptions } from "./auth";

const JWT_SECRET = process.env.JWT_SECRET || "your-secret-key";
const TOKEN_EXPIRY = "7d";
const COOKIE_NAME = "auth_token";

export interface SessionUser {
  id: string;
  email: string;
  username: string;
  role: string;
}

export function createSessionToken(user: SessionUser): string {
  return sign(
    {
      id: user.id,
      email: user.email,
      username: user.username,
      role: user.role,
    },
    JWT_SECRET,
    { expiresIn: TOKEN_EXPIRY }
  );
}

export function verifySessionToken(token: string): SessionUser | null {
  try {
    const decoded = verify(token, JWT_SECRET) as SessionUser;
    return decoded;
  } catch (error) {
    console.error("Token verification failed:", error);
    return null;
  }
}

export async function setSessionCookie(token: string) {
  const cookieStore = await cookies();
  cookieStore.set(COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: "/",
  });
}

export async function getSessionCookie(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get(COOKIE_NAME)?.value;
}

export async function clearSessionCookie() {
  const cookieStore = await cookies();
  cookieStore.delete(COOKIE_NAME);
}

export async function getCurrentUser(): Promise<SessionUser | null> {
  const token = await getSessionCookie();
  
  if (!token) {
    return null;
  }
  
  return verifySessionToken(token);
}

export async function requireAuth(): Promise<SessionUser> {
  const session = await getServerSession(authOptions);
  if (!session?.user) {
    throw new Error("Unauthorized");
  }
  
  return session.user as SessionUser;
}

export async function requireAdmin(): Promise<SessionUser> {
  const session = await getServerSession(authOptions);
  if(session==null){
    throw new Error("Unauthorized");
  }
  const user = session.user as AuthUser;
  
  console.log("requireAdmin",user)

  // if (user.role !== "admin") {
  //   throw new Error("Forbidden: Admin access required");
  // }
  
  return user;
}
