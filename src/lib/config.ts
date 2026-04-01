import prisma from "@/lib/prisma";

export interface SystemConfig {
  [key: string]: string;
}

interface ConfigCache {
  data: SystemConfig;
  timestamp: number;
}

const configCaches: Map<string, ConfigCache> = new Map();
const DEFAULT_CACHE_TTL = 5 * 60 * 1000;

export async function getConfigByCategory(
  category: string,
  cacheTTL: number = DEFAULT_CACHE_TTL
): Promise<SystemConfig> {
  const now = Date.now();
  const cached = configCaches.get(category);

  if (cached && (now - cached.timestamp) < cacheTTL) {
    return cached.data;
  }

  const settings = await prisma.systemSetting.findMany({
    where: {
      category,
    },
  });

  const config: SystemConfig = {};
  settings.forEach((s) => {
    config[s.key] = s.value;
  });

  configCaches.set(category, {
    data: config,
    timestamp: now,
  });

  return config;
}

export async function getConfigValue(
  category: string,
  key: string,
  defaultValue: string = ""
): Promise<string> {
  const config = await getConfigByCategory(category);
  return config[key] ?? defaultValue;
}

export async function getConfigValueAsBoolean(
  category: string,
  key: string,
  defaultValue: boolean = false
): Promise<boolean> {
  const value = await getConfigValue(category, key, defaultValue.toString());
  return value === "true";
}

export async function getConfigValueAsNumber(
  category: string,
  key: string,
  defaultValue: number = 0
): Promise<number> {
  const value = await getConfigValue(category, key, defaultValue.toString());
  const parsed = parseFloat(value);
  return isNaN(parsed) ? defaultValue : parsed;
}

export async function getConfigValueAsJSON<T = any>(
  category: string,
  key: string,
  defaultValue: T | null = null
): Promise<T | null> {
  const value = await getConfigValue(category, key);
  if (!value) {
    return defaultValue;
  }

  try {
    return JSON.parse(value) as T;
  } catch (error) {
    console.error(`Failed to parse JSON config for ${category}.${key}:`, error);
    return defaultValue;
  }
}

export function clearConfigCache(category?: string): void {
  if (category) {
    configCaches.delete(category);
  } else {
    configCaches.clear();
  }
}

export function setConfigCache(
  category: string,
  config: SystemConfig,
  cacheTTL: number = DEFAULT_CACHE_TTL
): void {
  configCaches.set(category, {
    data: config,
    timestamp: Date.now(),
  });
}

export function getCacheStats(): { category: string; age: number; size: number }[] {
  const now = Date.now();
  const stats: { category: string; age: number; size: number }[] = [];

  configCaches.forEach((cache, category) => {
    stats.push({
      category,
      age: now - cache.timestamp,
      size: Object.keys(cache.data).length,
    });
  });

  return stats;
}
