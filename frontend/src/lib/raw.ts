import type { ApiResponse } from "@/lib/api";

type Envelope = ApiResponse;

async function readEnvelope(response: Response): Promise<Envelope> {
  const payload: unknown = await response.json();
  if (typeof payload === "object" && payload !== null && typeof (payload as Record<string, unknown>).success === "boolean") {
    return payload as Envelope;
  }
  return { success: response.ok, error: response.ok ? null : "非标准响应" };
}

export async function rawGet<T = Envelope>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(path, { credentials: "include", ...init });
  return readEnvelope(response) as unknown as T;
}

export async function rawPost(
  path: string,
  body: Record<string, unknown>,
  init: RequestInit = {},
): Promise<Envelope> {
  const response = await fetch(path, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init.headers as Record<string, string> | undefined) },
    body: JSON.stringify(body),
    ...init,
  });
  return readEnvelope(response);
}

export async function uploadFile(file: File, path = "/api/upload"): Promise<Envelope> {
  const formData = new FormData();
  formData.append("file", file);
  const response = await fetch(path, {
    method: "POST",
    credentials: "include",
    body: formData,
  });
  return readEnvelope(response);
}
