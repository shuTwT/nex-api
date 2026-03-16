import { NextRequest, NextResponse } from "next/server";
import prisma from "@/lib/prisma";
import { hashPassword } from "@/lib/auth";
import { userCreateSchema, userUpdateSchema } from "@/lib/validations/user";

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const role = searchParams.get("role");
    const search = searchParams.get("search");
    const page = parseInt(searchParams.get("page") || "1");
    const limit = parseInt(searchParams.get("limit") || "10");
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
      password: undefined,
      subscription: user.subscriptions[0] || null,
      subscriptions: undefined,
    }));

    return NextResponse.json({
      success: true,
      data: formattedUsers,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    });
  } catch (error) {
    console.error("Error fetching users:", error);
    return NextResponse.json(
      { success: false, error: "Failed to fetch users" },
      { status: 500 }
    );
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    
    const validatedData = userCreateSchema.parse(body);
    
    const existingUser = await prisma.user.findFirst({
      where: {
        OR: [
          { email: validatedData.email },
          { username: validatedData.username },
        ],
      },
    });

    if (existingUser) {
      return NextResponse.json(
        {
          success: false,
          error: existingUser.email === validatedData.email
            ? "Email already exists"
            : "Username already exists",
        },
        { status: 400 }
      );
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

    return NextResponse.json({
      success: true,
      data: user,
    });
  } catch (error) {
    console.error("Error creating user:", error);
    
    if (error instanceof Error && error.name === "ZodError") {
      return NextResponse.json(
        { success: false, error: "Invalid input data", details: error },
        { status: 400 }
      );
    }

    return NextResponse.json(
      { success: false, error: "Failed to create user" },
      { status: 500 }
    );
  }
}

export async function PUT(request: NextRequest) {
  try {
    const body = await request.json();
    const { id, ...updateData } = body;

    if (!id) {
      return NextResponse.json(
        { success: false, error: "User ID is required" },
        { status: 400 }
      );
    }

    const validatedData = userUpdateSchema.parse(updateData);

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
        return NextResponse.json(
          {
            success: false,
            error: existingUser.email === validatedData.email
              ? "Email already exists"
              : "Username already exists",
          },
          { status: 400 }
        );
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

    return NextResponse.json({
      success: true,
      data: user,
    });
  } catch (error) {
    console.error("Error updating user:", error);
    
    if (error instanceof Error && error.name === "ZodError") {
      return NextResponse.json(
        { success: false, error: "Invalid input data", details: error },
        { status: 400 }
      );
    }

    return NextResponse.json(
      { success: false, error: "Failed to update user" },
      { status: 500 }
    );
  }
}

export async function DELETE(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const id = searchParams.get("id");

    if (!id) {
      return NextResponse.json(
        { success: false, error: "User ID is required" },
        { status: 400 }
      );
    }

    await prisma.user.delete({
      where: { id },
    });

    return NextResponse.json({
      success: true,
      message: "User deleted successfully",
    });
  } catch (error) {
    console.error("Error deleting user:", error);
    return NextResponse.json(
      { success: false, error: "Failed to delete user" },
      { status: 500 }
    );
  }
}
