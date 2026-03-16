"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import prisma from "@/lib/prisma";
import { hashPassword } from "@/lib/auth";
import { userCreateSchema, userUpdateSchema } from "@/lib/validations/user";
import { requireAdmin } from "@/lib/session";

export async function getUsers(params: {
  role?: string;
  search?: string;
  page?: number;
  limit?: number;
}) {
  try {
    const { role, search, page = 1, limit = 10 } = params;
    const skip = (page - 1) * limit;

    const where: any = {};

    if (role && role !== "all") {
      where.role = role;
    }

    if (search) {
      where.OR = [
        { username: { contains: search } },
        { email: { contains: search } },
      ];
    }

    const [users, total] = await Promise.all([
      prisma.user.findMany({
        where,
        select: {
          id: true,
          email: true,
          username: true,
          role: true,
          credits: true,
          createdAt: true,
          updatedAt: true,
          subscriptions: {
            where: { isActive: true },
            select: {
              plan: true,
              endDate: true,
            },
            take: 1,
          },
        },
        orderBy: { createdAt: "desc" },
        skip,
        take: limit,
      }),
      prisma.user.count({ where }),
    ]);

    const formattedUsers = users.map((user) => ({
      ...user,
      subscription: user.subscriptions[0] || null,
      subscriptions: undefined,
    }));

    return {
      success: true,
      data: formattedUsers,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  } catch (error) {
    console.error("Error fetching users:", error);
    return { success: false, error: "Failed to fetch users" };
  }
}

export async function getUserById(id: string) {
  try {
    const user = await prisma.user.findUnique({
      where: { id },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
        updatedAt: true,
        subscriptions: {
          where: { isActive: true },
          select: {
            id: true,
            plan: true,
            credits: true,
            price: true,
            startDate: true,
            endDate: true,
            isActive: true,
          },
          orderBy: { createdAt: "desc" },
          take: 1,
        },
        apiUsage: {
          select: {
            id: true,
            credits: true,
            status: true,
            createdAt: true,
            api: {
              select: {
                name: true,
                endpoint: true,
              },
            },
          },
          orderBy: { createdAt: "desc" },
          take: 10,
        },
      },
    });

    if (!user) {
      return { success: false, error: "User not found" };
    }

    return {
      success: true,
      data: {
        ...user,
        subscription: user.subscriptions[0] || null,
        subscriptions: undefined,
      },
    };
  } catch (error) {
    console.error("Error fetching user:", error);
    return { success: false, error: "Failed to fetch user" };
  }
}

export async function createUser(formData: FormData) {
  try {
    await requireAdmin();

    const data = {
      email: formData.get("email") as string,
      username: formData.get("username") as string,
      password: formData.get("password") as string,
      role: formData.get("role") as string,
      credits: parseInt(formData.get("credits") as string) || 1000,
    };

    const validatedData = userCreateSchema.parse(data);

    const existingUser = await prisma.user.findFirst({
      where: {
        OR: [
          { email: validatedData.email },
          { username: validatedData.username },
        ],
      },
    });

    if (existingUser) {
      return {
        success: false,
        error: existingUser.email === validatedData.email
          ? "邮箱已存在"
          : "用户名已存在",
      };
    }

    const hashedPassword = await hashPassword(validatedData.password);

    const user = await prisma.user.create({
      data: {
        email: validatedData.email,
        username: validatedData.username,
        password: hashedPassword,
        role: validatedData.role,
        credits: validatedData.credits,
      },
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    revalidatePath("/console/users");
    return { success: true, data: user };
  } catch (error) {
    console.error("Error creating user:", error);
    
    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "创建用户失败" };
  }
}

export async function updateUser(formData: FormData) {
  try {
    await requireAdmin();

    const id = formData.get("id") as string;
    const data = {
      email: formData.get("email") as string || undefined,
      username: formData.get("username") as string || undefined,
      role: formData.get("role") as string || undefined,
      credits: formData.get("credits") ? parseInt(formData.get("credits") as string) : undefined,
    };

    const validatedData = userUpdateSchema.parse(data);

    if (validatedData.email || validatedData.username) {
      const existingUser = await prisma.user.findFirst({
        where: {
          AND: [
            { id: { not: id } },
            {
              OR: [
                ...(validatedData.email ? [{ email: validatedData.email }] : []),
                ...(validatedData.username ? [{ username: validatedData.username }] : []),
              ],
            },
          ],
        },
      });

      if (existingUser) {
        return {
          success: false,
          error: existingUser.email === validatedData.email
            ? "邮箱已存在"
            : "用户名已存在",
        };
      }
    }

    const user = await prisma.user.update({
      where: { id },
      data: validatedData,
      select: {
        id: true,
        email: true,
        username: true,
        role: true,
        credits: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    revalidatePath("/console/users");
    return { success: true, data: user };
  } catch (error) {
    console.error("Error updating user:", error);
    
    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "更新用户失败" };
  }
}

export async function deleteUser(id: string) {
  try {
    await requireAdmin();

    await prisma.user.delete({
      where: { id },
    });

    revalidatePath("/console/users");
    return { success: true, message: "用户已删除" };
  } catch (error) {
    console.error("Error deleting user:", error);
    
    if (error instanceof Error && error.message === "Unauthorized") {
      redirect("/console");
    }

    return { success: false, error: "删除用户失败" };
  }
}

export async function getUserStats() {
  try {
    const [totalUsers, activeUsers, adminUsers, newUsersThisMonth] = await Promise.all([
      prisma.user.count(),
      prisma.user.count({
        where: {
          apiUsage: {
            some: {
              createdAt: {
                gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
              },
            },
          },
        },
      }),
      prisma.user.count({
        where: { role: "admin" },
      }),
      prisma.user.count({
        where: {
          createdAt: {
            gte: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
          },
        },
      }),
    ]);

    return {
      success: true,
      data: {
        totalUsers,
        activeUsers,
        adminUsers,
        newUsersThisMonth,
      },
    };
  } catch (error) {
    console.error("Error fetching user stats:", error);
    return { success: false, error: "获取用户统计失败" };
  }
}
