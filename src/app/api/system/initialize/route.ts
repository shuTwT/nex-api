import { NextRequest } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { hashPassword } from "@/lib/auth/util";
import { apiSuccess, apiError } from "@/lib/api-auth";

export async function POST(request: NextRequest) {
  try {
    const userCount = await prisma.user.count();
    if (userCount > 0) return apiError("系统已初始化，无法重复初始化", 400);

    const body = await request.json();
    const { email, username, password, confirmPassword } = body;

    if (!email || !username || !password || !confirmPassword) return apiError("请填写所有必填字段", 400);
    if (password !== confirmPassword) return apiError("两次输入的密码不一致", 400);
    if (password.length < 8) return apiError("密码长度至少为 8 位", 400);

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) return apiError("请输入有效的邮箱地址", 400);

    const hashedPw = await hashPassword(password);

    const user = await prisma.user.create({
      data: { email, username, password: hashedPw, role: "admin", credits: 10000 },
      select: { id: true, email: true, username: true, role: true },
    });

    revalidatePath("/");
    return apiSuccess({ ...user, password, credits: 10000 }, 201);
  } catch (error) {
    console.error("Initialize system error:", error);
    return apiError("初始化失败，请重试");
  }
}
