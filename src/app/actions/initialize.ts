"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { hashPassword } from "@/lib/auth/util";

export async function initializeSystem(formData: FormData): Promise<{
  success: boolean;
  error?: string;
}> {
  try {
    const userCount = await prisma.user.count();
    
    if (userCount > 0) {
      return {
        success: false,
        error: "系统已初始化，无法重复初始化",
      };
    }

    const email = formData.get("email") as string;
    const username = formData.get("username") as string;
    const password = formData.get("password") as string;
    const confirmPassword = formData.get("confirmPassword") as string;

    if (!email || !username || !password || !confirmPassword) {
      return {
        success: false,
        error: "请填写所有必填字段",
      };
    }

    if (password !== confirmPassword) {
      return {
        success: false,
        error: "两次输入的密码不一致",
      };
    }

    if (password.length < 8) {
      return {
        success: false,
        error: "密码长度至少为 8 位",
      };
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return {
        success: false,
        error: "请输入有效的邮箱地址",
      };
    }

    const hashedPassword = await hashPassword(password);

    const user = await prisma.user.create({
      data: {
        email,
        username,
        password: hashedPassword,
        role: "admin",
        credits: 10000,
      },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
      },
    });

    revalidatePath("/");
    return {
      success: true,
    };
  } catch (error) {
    console.error("Initialize system error:", error);
    return {
      success: false,
      error: "初始化失败，请重试",
    };
  }
}
