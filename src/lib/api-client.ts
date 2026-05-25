/* eslint-disable @typescript-eslint/no-explicit-any -- generic client wrapper defaults */
type ApiResponse<T = any> = {
  success: boolean;
  data?: T;
  error?: string;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  message?: string;
};

async function request<T = any>(
  url: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    const json = await res.json();
    return json as ApiResponse<T>;
  } catch (error) {
    console.error(`API request failed: ${url}`, error);
    return { success: false, error: "Network error" };
  }
}

function buildUrl(base: string, params?: Record<string, string | number | boolean | undefined | null>): string {
  if (!params) return base;
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== "") {
      searchParams.set(key, String(value));
    }
  }
  const qs = searchParams.toString();
  return qs ? `${base}?${qs}` : base;
}

export const api = {
  get<T = any>(url: string, params?: Record<string, string | number | boolean | undefined | null>): Promise<ApiResponse<T>> {
    return request<T>(buildUrl(url, params));
  },

  post<T = any>(url: string, body?: any): Promise<ApiResponse<T>> {
    return request<T>(url, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  put<T = any>(url: string, body?: any): Promise<ApiResponse<T>> {
    return request<T>(url, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  delete<T = any>(url: string, params?: Record<string, string | number | boolean | undefined | null>): Promise<ApiResponse<T>> {
    return request<T>(buildUrl(url, params), { method: "DELETE" });
  },

  paginated<T = any>(url: string, params?: Record<string, string | number | boolean | undefined | null>): Promise<ApiResponse<T>> {
    return this.get<T>(url, params);
  },

  async postFormData<T = any>(url: string, formData: FormData): Promise<ApiResponse<T>> {
    try {
      const res = await fetch(url, {
        method: "POST",
        body: formData,
      });
      return (await res.json()) as ApiResponse<T>;
    } catch (error) {
      console.error(`API form request failed: ${url}`, error);
      return { success: false, error: "Network error" };
    }
  },
};
