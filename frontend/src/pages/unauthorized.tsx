import { Link } from "react-router";
import { Button } from "@/components/ui/button";
import { Lock, ArrowLeft } from "lucide-react";

export default function UnauthorizedPage() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-center px-4">
        <div className="mb-8">
          <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-red-100 mb-6">
            <Lock className="h-12 w-12 text-red-600" />
          </div>
        </div>

        <h1 className="text-7xl font-bold text-slate-900 mb-4">
          401
        </h1>

        <h2 className="text-2xl font-semibold text-slate-700 mb-4">
          未授权访问
        </h2>

        <p className="text-lg text-slate-500 mb-8 max-w-md mx-auto">
          您需要登录后才能访问此页面。请登录您的账户以继续。
        </p>

        <div className="flex gap-4 justify-center">
          <Link to="/">
            <Button variant="outline" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              返回首页
            </Button>
          </Link>
          <Link to="/console">
            <Button className="bg-blue-600 hover:bg-blue-700">
              前往控制台
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}