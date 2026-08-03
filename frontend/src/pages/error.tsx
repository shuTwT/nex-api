import { useRouteError, Link, isRouteErrorResponse } from "react-router";
import { Button } from "@/components/ui/button";
import { AlertCircle, ArrowLeft, Home } from "lucide-react";

export function ErrorPage() {
  const error = useRouteError();
  const status = isRouteErrorResponse(error) ? error.status : 500;
  const message = isRouteErrorResponse(error)
    ? error.statusText || error.data
    : error instanceof Error
      ? error.message
      : "发生了意外错误";

  const isNotFound = status === 404;

  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-center px-4">
        <div className="mb-8">
          <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-red-100 mb-6">
            <AlertCircle className="h-12 w-12 text-red-600" />
          </div>
        </div>

        <h1 className="text-7xl font-bold text-slate-900 mb-4">
          {status}
        </h1>

        <h2 className="text-2xl font-semibold text-slate-700 mb-4">
          {isNotFound ? "页面未找到" : "出错了"}
        </h2>

        <p className="text-lg text-slate-500 mb-8 max-w-md mx-auto">
          {isNotFound
            ? "您访问的页面不存在或已被移动。"
            : typeof message === "string"
              ? message
              : "发生了意外错误，请稍后重试。"}
        </p>

        <div className="flex gap-4 justify-center">
          <Link to="/">
            <Button variant="outline" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              返回首页
            </Button>
          </Link>
          <Link to="/console">
            <Button className="bg-blue-600 hover:bg-blue-700 gap-2">
              <Home className="h-4 w-4" />
              前往控制台
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}

export default ErrorPage;