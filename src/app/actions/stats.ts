"use server";

import { authOptions } from "@/lib/auth/config";
import { requireAuth } from "@/lib/auth/util";
import {
  getGlobalRequestCount,
  getApiRequestCount,
  getUserApiRequestCount,
  getAllApiRequestCounts,
} from "@/lib/request-stats";

export async function getGlobalStats() {
  try {
    await requireAuth(authOptions);
    const count = await getGlobalRequestCount();
    return { success: true, data: { totalRequests: count } };
  } catch (error) {
    console.error("Error getting global stats:", error);
    return { success: false, error: "获取全局统计失败" };
  }
}

export async function getApiStats(apiAlias: string) {
  try {
    await requireAuth(authOptions);
    const count = await getApiRequestCount(apiAlias);
    return { success: true, data: { apiAlias, requestCount: count } };
  } catch (error) {
    console.error("Error getting API stats:", error);
    return { success: false, error: "获取 API 统计失败" };
  }
}

export async function getUserApiStats(apiAlias: string) {
  try {
    const user = await requireAuth(authOptions);
    const count = await getUserApiRequestCount(user.id, apiAlias);
    return { success: true, data: { apiAlias, requestCount: count } };
  } catch (error) {
    console.error("Error getting user API stats:", error);
    return { success: false, error: "获取用户 API 统计失败" };
  }
}

export async function getAllApiStats() {
  try {
    await requireAuth(authOptions);
    const stats = await getAllApiRequestCounts();
    return { success: true, data: stats };
  } catch (error) {
    console.error("Error getting all API stats:", error);
    return { success: false, error: "获取所有 API 统计失败" };
  }
}
