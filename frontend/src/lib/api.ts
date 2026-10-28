import createClient from "openapi-fetch";

import type { components, operations as operationSchemas, paths } from "@/api/generated/schema";
import { operations } from "@/api/generated/operations";
import type { OperationName } from "@/api/generated/operations";
import { config } from "@/lib/config";

/** Response code meaning success; every other value is an entry of the backend error code table. */
export const CODE_OK = 0;

/** Uniform response body: `{ code, msg, data }`. HTTP statuses stay real; `code` refines them. */
export type ApiResponse = {
  code: number;
  msg: string;
  data?: unknown;
};

/** The `data` payload of a paginated endpoint. */
export type PageData<T> = {
  list: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
};

export function isOk(res: ApiResponse): boolean {
  return res.code === CODE_OK;
}

type JsonObject = Record<string, unknown>;

/** The type-safe client for new API calls. It is backed by openapi-typescript generated paths. */
export const typedApi = createClient<paths>({
  baseUrl: config.apiUrl,
  credentials: "include",
});

// ---------------------------------------------------------------------------
// Typed operation facade
//
// `operations` (generated from openapi.yaml by tools/genoperations) supplies
// the operationId -> [method, path] mapping at runtime; `operationSchemas`
// (generated from the same document) supplies each operation's parameter and
// body types. Together they make unknown or misspelled operation ids a compile
// error and validate path, query and JSON body payloads against the spec.
// ---------------------------------------------------------------------------

type OpParameters<O extends OperationName> = NonNullable<
  operationSchemas[O] extends { parameters?: infer Parameters } ? Parameters : never
>;

type PathParams<O extends OperationName> =
  OpParameters<O> extends { path?: infer Path }
    ? Exclude<Path, undefined> extends Record<string, never>
      ? Record<string, never>
      : Exclude<Path, undefined>
    : Record<string, never>;

type QueryParams<O extends OperationName> =
  OpParameters<O> extends { query?: infer Query }
    ? Exclude<Query, undefined> extends Record<string, never>
      ? Record<string, never>
      : Exclude<Query, undefined>
    : Record<string, never>;

type RequestBody<O extends OperationName> =
  operationSchemas[O] extends { requestBody?: { content: { "application/json": infer Body } } }
    ? Body
    : never;

type OpMethod<O extends OperationName> = (typeof operations)[O][0];
type ReadsPayload<O extends OperationName> = OpMethod<O> extends "get" | "delete" | "options" ? true : false;
type Payload<O extends OperationName> = ReadsPayload<O> extends true ? QueryParams<O> : RequestBody<O>;

type HasPathParams<O extends OperationName> = PathParams<O> extends Record<string, never> ? false : true;

type ApiFn<O extends OperationName> = HasPathParams<O> extends true
  ? (pathParams: PathParams<O>, payload?: Payload<O>) => Promise<ApiResponse>
  : (payload?: Payload<O>) => Promise<ApiResponse>;

export type ApiClient = { [O in OperationName]: ApiFn<O> };

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isEnvelope(value: unknown): value is ApiResponse {
  return isObject(value) && typeof value.code === "number";
}

function expandPath(template: string, parameters: JsonObject | undefined): string {
  return template.replace(/\{([A-Za-z0-9_]+)\}/g, (_match, name: string) => {
    const value = parameters?.[name];
    if (value === undefined || value === null) throw new Error("Missing path parameter: " + name);
    return encodeURIComponent(String(value));
  });
}

async function invoke(operation: OperationName, args: unknown[]): Promise<ApiResponse> {
  const [method, template] = operations[operation];
  const pathParameters = template.includes("{") && isObject(args[0]) ? (args.shift() as JsonObject) : undefined;
  const path = expandPath(template, pathParameters);
  const payload = args[0];
  const options = method === "get" || method === "delete" || method === "options"
    ? { params: payload === undefined ? undefined : { query: payload } }
    : { body: payload };
  const response = await (typedApi as unknown as Record<string, (route: string, options: unknown) => Promise<{ data?: unknown; error?: unknown }>>)[method.toUpperCase()](path, options);
  if (isEnvelope(response.data)) return response.data;
  if (isEnvelope(response.error)) return response.error;
  return { code: -1, msg: "API request failed" };
}

/**
 * Backward-compatible facade for existing pages. It delegates all I/O to typedApi.
 * Prefer typedApi for newly written code.
 */
export const api: ApiClient = new Proxy({} as ApiClient, {
  get: (_target, property) => {
    if (typeof property !== "string" || !(property in operations)) return undefined;
    return (...args: unknown[]) => invoke(property as OperationName, args);
  },
});

export function unwrap<T>(res: ApiResponse): T {
  if (!isOk(res)) throw new Error(res.msg || "API request failed");
  return (res.data ?? null) as T;
}

export function responseData<T>(res: ApiResponse): T | null {
  return isOk(res) ? ((res.data ?? null) as T) : null;
}

/** Returns the paginated payload of a successful response, or null. */
export function pageData<T>(res: ApiResponse): PageData<T> | null {
  return isOk(res) ? ((res.data ?? null) as PageData<T> | null) : null;
}
