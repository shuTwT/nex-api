import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router";
import { api, responseData } from "@/lib/api";

type InitializationState =
  | { status: "checking" }
  | { status: "ready"; initialized: boolean }
  | { status: "error" };

export function InitializationGuard() {
  const location = useLocation();
  const [retryCount, setRetryCount] = useState(0);
  const [state, setState] = useState<InitializationState>({
    status: "checking",
  });

  useEffect(() => {
    let cancelled = false;

    async function checkInitialization() {
      setState({ status: "checking" });
      try {
        const result = await api.system_initialized_route_get();
        const data = responseData<{ initialized: boolean }>(result);
        if (!data || typeof data.initialized !== "boolean") {
          throw new Error("Invalid initialization status response");
        }
        if (!cancelled) {
          setState({ status: "ready", initialized: data.initialized });
        }
      } catch {
        if (!cancelled) {
          setState({ status: "error" });
        }
      }
    }

    void checkInitialization();
    return () => {
      cancelled = true;
    };
  }, [retryCount]);

  if (state.status === "checking") {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (state.status === "error") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 p-4">
        <div className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-6 text-center shadow-sm">
          <h1 className="text-lg font-semibold text-slate-900">
            无法检查系统初始化状态
          </h1>
          <p className="mt-2 text-sm text-slate-500">
            请确认后端服务可用后重试。
          </p>
          <button
            type="button"
            className="mt-5 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            onClick={() => setRetryCount((count) => count + 1)}
          >
            重新检查
          </button>
        </div>
      </div>
    );
  }

  const normalizedPath = location.pathname.replace(/\/+$/, "") || "/";
  const isInitializePage = normalizedPath === "/initialize";

  if (!state.initialized && !isInitializePage) {
    return <Navigate to="/initialize" replace />;
  }

  if (state.initialized && isInitializePage) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
