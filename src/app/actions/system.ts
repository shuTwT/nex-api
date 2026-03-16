"use server";

import prisma from "@/lib/prisma";

export async function checkSystemInitialized() {
  try {
    const userCount = await prisma.user.count();
    return {
      initialized: userCount > 0,
    };
  } catch (error) {
    console.error("Check system initialized error:", error);
    return {
      initialized: false,
    };
  }
}
