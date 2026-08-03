import { NexApiClient, type ApiResponse } from "@/api/generated";
import { config } from "@/lib/config";

/**
 * Shared API client singleton for same-origin calls.
 *
 * All requests go to the same origin (`/api/**`) — in dev the Vite proxy
 * forwards `/api` to the Go backend; in production the Go binary serves
 * the SPA and the API from one origin, so no CORS is involved.
 */
export const api = new NexApiClient({
  baseUrl: config.apiUrl,
  credentials: "include",
});

/**
 * Unwrap an ApiResponse or throw on failure.
 * Convenience for call sites that want `data` directly.
 */
export function unwrap<T>(res: ApiResponse): T {
  if (!res.success) {
    throw new Error(res.error ?? "API request failed");
  }
  return (res.data ?? null) as T;
}

export function responseData<T>(res: ApiResponse): T | null {
  return res.success ? ((res.data ?? null) as T) : null;
}
