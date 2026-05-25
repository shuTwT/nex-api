import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { getAdminUser, apiSuccess, apiError, apiPaginated } from "@/lib/api-auth";
import { hashPassword } from "@/lib/auth/util";
import { userCreateSchema } from "@/lib/validations/user";

export async function GET(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  const { searchParams } = new URL(request.url);
  const role = searchParams.get("role") || undefined;
  const search = searchParams.get("search") || undefined;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "10");
  const skip = (page - 1) * limit;

  const where: Record<string, unknown> = {};
  if (role && role !== "all") where.role = role;
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
        id: true, email: true, username: true, role: true, credits: true,
        createdAt: true, updatedAt: true,
        subscriptions: { where: { isActive: true }, select: { planName: true, endDate: true }, take: 1 },
      },
      orderBy: { createdAt: "desc" },
      skip, take: limit,
    }),
    prisma.user.count({ where }),
  ]);

  const formattedUsers = users.map((user) => ({
    ...user,
    subscription: user.subscriptions[0] || null,
    subscriptions: undefined,
  }));

  return apiPaginated(formattedUsers, { page, limit, total });
}

export async function POST(request: NextRequest) {
  const user = await getAdminUser();
  if (user instanceof NextResponse) return user;

  try {
    const body = await request.json();
    const validatedData = userCreateSchema.parse(body);

    const existingUser = await prisma.user.findFirst({
      where: { OR: [{ email: validatedData.email }, { username: validatedData.username }] },
    });

    if (existingUser) {
      return apiError(existingUser.email === validatedData.email ? "邮箱已存在" : "用户名已存在", 409);
    }

    const hashedPassword = await hashPassword(validatedData.password);
    const newUser = await prisma.user.create({
      data: { ...validatedData, password: hashedPassword },
      select: { id: true, email: true, username: true, role: true, credits: true, createdAt: true, updatedAt: true },
    });

    revalidatePath("/console/users");
    return apiSuccess(newUser, 201);
  } catch (error) {
    console.error("Error creating user:", error);
    return apiError("创建用户失败");
  }
}
