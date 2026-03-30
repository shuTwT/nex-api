import redis from "./redis";

const KEYS = {
  GLOBAL_REQUEST_COUNT: "global:request:count",
  API_REQUEST_COUNT: (apiAlias: string) => `api:request:count:${apiAlias}`,
  USER_API_REQUEST_COUNT: (userId: string, apiAlias: string) =>
    `user:api:request:count:${userId}:${apiAlias}`,
  GLOBAL_HOURLY_USAGE: (timestamp: number) => `global:usage:hourly:${timestamp}`,
  USER_HOURLY_USAGE: (userId: string, timestamp: number) =>
    `user:usage:hourly:${userId}:${timestamp}`,
};

function getHourTimestamp(date: Date = new Date()): number {
  const d = new Date(date);
  d.setMinutes(0, 0, 0);
  return Math.floor(d.getTime() / 1000);
}

export async function incrementRequestCount(
  userId: string,
  apiAlias: string,
  credits: number = 0
): Promise<void> {
  try {
    const pipeline = redis.pipeline();
    const hourTimestamp = getHourTimestamp();

    pipeline.incr(KEYS.GLOBAL_REQUEST_COUNT);
    pipeline.incr(KEYS.API_REQUEST_COUNT(apiAlias));
    pipeline.incr(KEYS.USER_API_REQUEST_COUNT(userId, apiAlias));

    if (credits > 0) {
      pipeline.incrbyfloat(KEYS.GLOBAL_HOURLY_USAGE(hourTimestamp), credits);
      pipeline.incrbyfloat(KEYS.USER_HOURLY_USAGE(userId, hourTimestamp), credits);
    }

    await pipeline.exec();
  } catch (error) {
    console.error("Error incrementing request count:", error);
  }
}

export async function getGlobalRequestCount(): Promise<number> {
  try {
    const count = await redis.get(KEYS.GLOBAL_REQUEST_COUNT);
    return parseInt(count || "0", 10);
  } catch (error) {
    console.error("Error getting global request count:", error);
    return 0;
  }
}

export async function getApiRequestCount(apiAlias: string): Promise<number> {
  try {
    const count = await redis.get(KEYS.API_REQUEST_COUNT(apiAlias));
    return parseInt(count || "0", 10);
  } catch (error) {
    console.error("Error getting API request count:", error);
    return 0;
  }
}

export async function getUserApiRequestCount(
  userId: string,
  apiAlias: string
): Promise<number> {
  try {
    const count = await redis.get(KEYS.USER_API_REQUEST_COUNT(userId, apiAlias));
    return parseInt(count || "0", 10);
  } catch (error) {
    console.error("Error getting user API request count:", error);
    return 0;
  }
}

export async function getAllApiRequestCounts(): Promise<
  Record<string, number>
> {
  try {
    const keys = await redis.keys("api:request:count:*");
    if (keys.length === 0) {
      return {};
    }

    const values = await redis.mget(...keys);
    const result: Record<string, number> = {};

    keys.forEach((key, index) => {
      const apiAlias = key.replace("api:request:count:", "");
      result[apiAlias] = parseInt(values[index] || "0", 10);
    });

    return result;
  } catch (error) {
    console.error("Error getting all API request counts:", error);
    return {};
  }
}

export async function getAllUserApiRequestCounts(): Promise<
  Record<string, number>
> {
  try {
    const keys = await redis.keys("user:api:request:count:*");
    if (keys.length === 0) {
      return {};
    }

    const values = await redis.mget(...keys);
    const result: Record<string, number> = {};

    keys.forEach((key, index) => {
      const keyWithoutPrefix = key.replace("user:api:request:count:", "");
      result[keyWithoutPrefix] = parseInt(values[index] || "0", 10);
    });

    return result;
  } catch (error) {
    console.error("Error getting all user API request counts:", error);
    return {};
  }
}

export async function syncToDatabase(): Promise<void> {
  try {
    const {prisma} = await import("@/lib/prisma");

    const globalCount = await getGlobalRequestCount();
    const apiCounts = await getAllApiRequestCounts();
    const userApiCounts = await getAllUserApiRequestCounts();

    console.log("Syncing Redis data to database...");
    console.log(`Global count: ${globalCount}`);
    console.log(`API counts: ${Object.keys(apiCounts).length} APIs`);
    console.log(`User-API counts: ${Object.keys(userApiCounts).length} records`);

    for (const [apiAlias, count] of Object.entries(apiCounts)) {
      await prisma.api.update({
        where: { alias: apiAlias },
        data: { callCount: count },
      }).catch((error) => {
        console.error(`Error updating API ${apiAlias}:`, error);
      });
    }

    for (const [key, count] of Object.entries(userApiCounts)) {
      const [userId, apiAlias] = key.split(":");
      if (userId && apiAlias) {
        await prisma.apiUsage.updateMany({
          where: {
            userId,
            api: { alias: apiAlias },
          },
          data: {
            credits: count,
          },
        }).catch((error) => {
          console.error(`Error updating user API usage ${key}:`, error);
        });
      }
    }

    await prisma.$disconnect();
    console.log("Sync completed successfully");
  } catch (error) {
    console.error("Error syncing to database:", error);
  }
}

export async function getHourlyUsageTrend(
  userId?: string,
  hours: number = 7
): Promise<number[]> {
  try {
    const now = new Date();
    const timestamps: number[] = [];
    
    for (let i = hours - 1; i >= 0; i--) {
      const d = new Date(now);
      d.setHours(d.getHours() - i);
      timestamps.push(getHourTimestamp(d));
    }

    const keys = userId
      ? timestamps.map((ts) => KEYS.USER_HOURLY_USAGE(userId, ts))
      : timestamps.map((ts) => KEYS.GLOBAL_HOURLY_USAGE(ts));

    const values = await redis.mget(...keys);
    
    return values.map((v) => parseFloat(v || "0"));
  } catch (error) {
    console.error("Error getting hourly usage trend:", error);
    return new Array(hours).fill(0);
  }
}
