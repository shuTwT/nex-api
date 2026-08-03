/**
 * Public runtime configuration sourced from VITE_ environment variables.
 * These values are exposed to the browser — never put server secrets here.
 */

function getEnv(key: keyof ImportMetaEnv, fallback = ""): string {
  const value = import.meta.env[key];
  return typeof value === "string" ? value : fallback;
}

export const config = {
  /** API base URL. Empty in dev (Vite proxy forwards /api to the Go backend). */
  apiUrl: getEnv("VITE_API_BASE_URL"),
  /** Public application URL. */
  appUrl: getEnv("VITE_APP_URL", "http://localhost:3000"),
} as const;

export type AppConfig = typeof config;