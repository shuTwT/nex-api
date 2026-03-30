"use server";

import { revalidatePath } from "next/cache";
import prisma from "@/lib/prisma";
import { requireAdmin } from "@/lib/session";

export async function getSubscriptionPlans() {
  try {
    const plans = await prisma.subscriptionPlan.findMany({
      orderBy: { sortOrder: "asc" },
    });

    return { success: true, data: plans };
  } catch (error) {
    console.error("Error fetching subscription plans:", error);
    return { success: false, error: "获取订阅计划失败" };
  }
}

export async function getSubscriptionPlanById(id: string) {
  try {
    const plan = await prisma.subscriptionPlan.findUnique({
      where: { id },
    });

    if (!plan) {
      return { success: false, error: "订阅计划未找到" };
    }

    return { success: true, data: plan };
  } catch (error) {
    console.error("Error fetching subscription plan:", error);
    return { success: false, error: "获取订阅计划失败" };
  }
}

export async function createSubscriptionPlan(formData: FormData) {
  try {
    await requireAdmin();

    const data = {
      title: formData.get("title") as string,
      price: parseFloat(formData.get("price") as string) || 0,
      totalCredits: parseInt(formData.get("totalCredits") as string) || 0,
      sortOrder: parseInt(formData.get("sortOrder") as string) || 0,
      validityDuration: parseInt(formData.get("validityDuration") as string) || 30,
      validityUnit: formData.get("validityUnit") as string || "day",
      creditResetCycle: formData.get("creditResetCycle") as string || "month",
      isActive: (formData.get("isActive") as string) === "on",
    };

    const plan = await prisma.subscriptionPlan.create({
      data,
    });

    revalidatePath("/console/subscription-plans");
    return { success: true, data: plan };
  } catch (error) {
    console.error("Error creating subscription plan:", error);
    return { success: false, error: "创建订阅计划失败" };
  }
}

export async function updateSubscriptionPlan(formData: FormData) {
  try {
    await requireAdmin();

    const id = formData.get("id") as string;
    const data = {
      title: formData.get("title") as string || undefined,
      price: formData.get("price") ? parseFloat(formData.get("price") as string) : undefined,
      totalCredits: formData.get("totalCredits") ? parseInt(formData.get("totalCredits") as string) : undefined,
      sortOrder: formData.get("sortOrder") ? parseInt(formData.get("sortOrder") as string) : undefined,
      validityDuration: formData.get("validityDuration") ? parseInt(formData.get("validityDuration") as string) : undefined,
      validityUnit: formData.get("validityUnit") as string || undefined,
      creditResetCycle: formData.get("creditResetCycle") as string || undefined,
      isActive: formData.get("isActive") ? (formData.get("isActive") as string) === "on" : undefined,
    };

    const plan = await prisma.subscriptionPlan.update({
      where: { id },
      data,
    });

    revalidatePath("/console/subscription-plans");
    return { success: true, data: plan };
  } catch (error) {
    console.error("Error updating subscription plan:", error);
    return { success: false, error: "更新订阅计划失败" };
  }
}

export async function deleteSubscriptionPlan(id: string) {
  try {
    await requireAdmin();

    await prisma.subscriptionPlan.delete({
      where: { id },
    });

    revalidatePath("/console/subscription-plans");
    return { success: true, message: "订阅计划已删除" };
  } catch (error) {
    console.error("Error deleting subscription plan:", error);
    return { success: false, error: "删除订阅计划失败" };
  }
}
