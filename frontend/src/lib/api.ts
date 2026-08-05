import createClient from "openapi-fetch";

import type { components, paths } from "@/api/generated/schema";
import { config } from "@/lib/config";

type GeneratedEnvelope = components["schemas"]["main.SwaggerEnvelope"];

// Swaggo's schema marks JSON fields optional; the HTTP envelope guarantees
// success and uses the following stable pagination shape at runtime.
export type ApiResponse = Omit<GeneratedEnvelope, "success" | "data" | "error" | "pagination"> & {
  success: boolean;
  data?: unknown;
  error?: string | null;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  } | null;
};

type ApiMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "OPTIONS";
type LegacyOperation = readonly [ApiMethod, string];
type JsonObject = Record<string, unknown>;

/** The type-safe client for new API calls. It is backed by openapi-typescript generated paths. */
export const typedApi = createClient<paths>({
  baseUrl: config.apiUrl,
  credentials: "include",
});

// Compatibility table for existing UI call sites. New code should use typedApi.GET/POST/etc.
const legacyOperations: Record<string, LegacyOperation> = {
  "advertisements_id_route_get": ["GET", "/api/advertisements/{id}"],
  "advertisements_id_route_put": ["PUT", "/api/advertisements/{id}"],
  "advertisements_id_route_delete": ["DELETE", "/api/advertisements/{id}"],
  "advertisements_id_toggle_route_put": ["PUT", "/api/advertisements/{id}/toggle"],
  "advertisements_by_position_position_route_get": ["GET", "/api/advertisements/by-position/{position}"],
  "advertisements_route_get": ["GET", "/api/advertisements"],
  "advertisements_route_post": ["POST", "/api/advertisements"],
  "advertisements_stats_route_get": ["GET", "/api/advertisements/stats"],
  "apis_id_route_get": ["GET", "/api/apis/{id}"],
  "apis_id_route_put": ["PUT", "/api/apis/{id}"],
  "apis_id_route_delete": ["DELETE", "/api/apis/{id}"],
  "apis_id_toggle_route_put": ["PUT", "/api/apis/{id}/toggle"],
  "apis_route_get": ["GET", "/api/apis"],
  "apis_route_post": ["POST", "/api/apis"],
  "apis_stats_route_get": ["GET", "/api/apis/stats"],
  "audit_logs_id_route_put": ["PUT", "/api/audit-logs/{id}"],
  "audit_logs_id_route_delete": ["DELETE", "/api/audit-logs/{id}"],
  "audit_logs_export_route_get": ["GET", "/api/audit-logs/export"],
  "audit_logs_route_get": ["GET", "/api/audit-logs"],
  "audit_logs_route_post": ["POST", "/api/audit-logs"],
  "audit_logs_stats_route_get": ["GET", "/api/audit-logs/stats"],
  "auth_logout_route_post": ["POST", "/api/auth/logout"],
  "auth_me_route_get": ["GET", "/api/auth/me"],
  "categories_id_route_put": ["PUT", "/api/categories/{id}"],
  "categories_id_route_delete": ["DELETE", "/api/categories/{id}"],
  "categories_route_get": ["GET", "/api/categories"],
  "categories_route_post": ["POST", "/api/categories"],
  "cron_sync_stats_route_post": ["POST", "/api/cron/sync-stats"],
  "dashboard_activity_route_get": ["GET", "/api/dashboard/activity"],
  "dashboard_stats_route_get": ["GET", "/api/dashboard/stats"],
  "dashboard_top_apis_route_get": ["GET", "/api/dashboard/top-apis"],
  "dashboard_usage_trend_route_get": ["GET", "/api/dashboard/usage-trend"],
  "marketplace_apis_id_route_get": ["GET", "/api/marketplace/apis/{id}"],
  "marketplace_apis_route_get": ["GET", "/api/marketplace/apis"],
  "marketplace_mcp_services_route_get": ["GET", "/api/marketplace/mcp-services"],
  "marketplace_mcp_services_id_route_get": ["GET", "/api/marketplace/mcp-services/{id}"],
  "marketplace_mcp_services_id_tools_route_get": ["GET", "/api/marketplace/mcp-services/{id}/tools"],
  "marketplace_mcp_stats_route_get": ["GET", "/api/marketplace/mcp-stats"],
  "marketplace_stats_route_get": ["GET", "/api/marketplace/stats"],
  "mcp_services_id_route_put": ["PUT", "/api/mcp-services/{id}"],
  "mcp_services_id_route_delete": ["DELETE", "/api/mcp-services/{id}"],
  "mcp_services_id_toggle_route_put": ["PUT", "/api/mcp-services/{id}/toggle"],
  "mcp_services_route_get": ["GET", "/api/mcp-services"],
  "mcp_services_route_post": ["POST", "/api/mcp-services"],
  "mcp_services_stats_route_get": ["GET", "/api/mcp-services/stats"],
  "membership_current_route_get": ["GET", "/api/membership/current"],
  "membership_plans_route_get": ["GET", "/api/membership/plans"],
  "membership_subscribe_route_post": ["POST", "/api/membership/subscribe"],
  "payment_outtradeno_cancel_route_post": ["POST", "/api/payment/{outTradeNo}/cancel"],
  "payment_outtradeno_route_get": ["GET", "/api/payment/{outTradeNo}"],
  "payment_outtradeno_status_route_get": ["GET", "/api/payment/{outTradeNo}/status"],
  "payment_business_recharge_route_post": ["POST", "/api/payment/business/recharge"],
  "payment_business_subscription_route_post": ["POST", "/api/payment/business/subscription"],
  "payment_callback_alipay_route_post": ["POST", "/api/payment/callback/alipay"],
  "payment_callback_mock_route_post": ["POST", "/api/payment/callback/mock"],
  "payment_callback_wechat_route_post": ["POST", "/api/payment/callback/wechat"],
  "payment_methods_route_get": ["GET", "/api/payment/methods"],
  "payment_methods_route_post": ["POST", "/api/payment/methods"],
  "payment_settings_route_get": ["GET", "/api/payment/settings"],
  "payment_user_route_get": ["GET", "/api/payment/user"],
  "personal_profile_route_get": ["GET", "/api/personal/profile"],
  "personal_redeem_lookup_route_post": ["POST", "/api/personal/redeem/lookup"],
  "personal_redeem_route_post": ["POST", "/api/personal/redeem"],
  "recharge_route_post": ["POST", "/api/recharge"],
  "redemption_codes_id_route_delete": ["DELETE", "/api/redemption-codes/{id}"],
  "redemption_codes_batch_route_delete": ["DELETE", "/api/redemption-codes/batch"],
  "redemption_codes_batch_route_post": ["POST", "/api/redemption-codes/batch"],
  "redemption_codes_export_route_get": ["GET", "/api/redemption-codes/export"],
  "redemption_codes_plans_route_get": ["GET", "/api/redemption-codes/plans"],
  "redemption_codes_route_get": ["GET", "/api/redemption-codes"],
  "redemption_codes_route_post": ["POST", "/api/redemption-codes"],
  "scheduled_jobs_id_route_delete": ["DELETE", "/api/scheduled-jobs/{id}"],
  "scheduled_jobs_id_route_get": ["GET", "/api/scheduled-jobs/{id}"],
  "scheduled_jobs_id_route_put": ["PUT", "/api/scheduled-jobs/{id}"],
  "scheduled_jobs_id_run_route_post": ["POST", "/api/scheduled-jobs/{id}/run"],
  "scheduled_jobs_route_get": ["GET", "/api/scheduled-jobs"],
  "scheduled_jobs_route_post": ["POST", "/api/scheduled-jobs"],
  "scheduled_jobs_tasks_route_get": ["GET", "/api/scheduled-jobs/tasks"],
  "stats_alias_route_get": ["GET", "/api/stats/{alias}"],
  "stats_route_get": ["GET", "/api/stats"],
  "subscription_plans_id_route_get": ["GET", "/api/subscription-plans/{id}"],
  "subscription_plans_id_route_put": ["PUT", "/api/subscription-plans/{id}"],
  "subscription_plans_id_route_delete": ["DELETE", "/api/subscription-plans/{id}"],
  "subscription_plans_route_get": ["GET", "/api/subscription-plans"],
  "subscription_plans_route_post": ["POST", "/api/subscription-plans"],
  "system_initialize_route_post": ["POST", "/api/system/initialize"],
  "system_initialized_route_get": ["GET", "/api/system/initialized"],
  "system_settings_announcement_route_get": ["GET", "/api/system-settings/announcement"],
  "system_settings_defaults_route_get": ["GET", "/api/system-settings/defaults"],
  "system_settings_route_get": ["GET", "/api/system-settings"],
  "system_settings_route_put": ["PUT", "/api/system-settings"],
  "tokens_id_route_put": ["PUT", "/api/tokens/{id}"],
  "tokens_id_route_delete": ["DELETE", "/api/tokens/{id}"],
  "tokens_id_toggle_route_put": ["PUT", "/api/tokens/{id}/toggle"],
  "tokens_route_get": ["GET", "/api/tokens"],
  "tokens_route_post": ["POST", "/api/tokens"],
  "tokens_stats_route_get": ["GET", "/api/tokens/stats"],
  "upload_filename_route_get": ["GET", "/api/upload/{filename}"],
  "upload_route_post": ["POST", "/api/upload"],
  "usage_route_get": ["GET", "/api/usage"],
  "users_id_route_get": ["GET", "/api/users/{id}"],
  "users_id_route_put": ["PUT", "/api/users/{id}"],
  "users_id_route_delete": ["DELETE", "/api/users/{id}"],
  "users_route_get": ["GET", "/api/users"],
  "users_route_post": ["POST", "/api/users"],
  "users_stats_route_get": ["GET", "/api/users/stats"],
  "v1_alias_route_get": ["GET", "/api/v1/{alias}"],
  "v1_alias_route_post": ["POST", "/api/v1/{alias}"],
  "v1_alias_route_put": ["PUT", "/api/v1/{alias}"],
  "v1_alias_route_delete": ["DELETE", "/api/v1/{alias}"],
  "v1_alias_route_patch": ["PATCH", "/api/v1/{alias}"],
  "v1_mcp_identifier_route_options": ["OPTIONS", "/api/v1/mcp/{identifier}"],
  "v1_mcp_identifier_route_post": ["POST", "/api/v1/mcp/{identifier}"],
};

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isEnvelope(value: unknown): value is ApiResponse {
  return isObject(value) && typeof value.success === "boolean";
}

function isPathParameters(value: unknown): value is JsonObject {
  return isObject(value);
}

function expandPath(template: string, parameters: JsonObject | undefined): string {
  return template.replace(/\{([A-Za-z0-9_]+)\}/g, (_match, name: string) => {
    const value = parameters?.[name];
    if (value === undefined || value === null) throw new Error("Missing path parameter: " + name);
    return encodeURIComponent(String(value));
  });
}

async function invoke(operation: string, args: unknown[]): Promise<ApiResponse> {
  const descriptor = legacyOperations[operation];
  if (!descriptor) throw new Error("Unknown API operation: " + operation);

  const [method, template] = descriptor;
  const pathParameters = template.includes("{") && isPathParameters(args[0]) ? (args.shift() as JsonObject) : undefined;
  const path = expandPath(template, pathParameters);
  const payload = args[0];
  const options = method === "GET" || method === "DELETE" || method === "OPTIONS"
    ? { params: payload === undefined ? undefined : { query: payload } }
    : { body: payload };
  const response = await (typedApi as unknown as Record<string, (route: string, options: unknown) => Promise<{ data?: unknown; error?: unknown }>>)[method](path, options);
  if (isEnvelope(response.data)) return response.data;
  if (isEnvelope(response.error)) return response.error;
  return { success: false, error: "API request failed" };
}

/**
 * Backward-compatible facade for existing pages. It delegates all I/O to typedApi.
 * Prefer typedApi for newly written code.
 */
export const api = new Proxy({} as Record<string, (...args: unknown[]) => Promise<ApiResponse>>, {
  get: (_target, property) => typeof property === "string" ? (...args: unknown[]) => invoke(property, args) : undefined,
});

export function unwrap<T>(res: ApiResponse): T {
  if (!res.success) throw new Error(res.error ?? "API request failed");
  return (res.data ?? null) as T;
}

export function responseData<T>(res: ApiResponse): T | null {
  return res.success ? ((res.data ?? null) as T) : null;
}
