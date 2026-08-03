import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { AlertCircle, ArrowLeft, Home } from "lucide-react";
import { useSearchParams, useNavigate, Link } from "react-router";

const errorMessages: Record<string, { title: string; description: string }> = {
  Configuration: {
    title: "配置错误",
    description: "服务器配置有问题，请联系管理员。",
  },
  AccessDenied: {
    title: "访问被拒绝",
    description: "您没有权限访问此资源。",
  },
  Verification: {
    title: "验证失败",
    description: "验证链接无效或已过期，请重新尝试。",
  },
  Default: {
    title: "登录错误",
    description: "登录过程中发生了错误，请稍后重试。",
  },
};

export default function AuthErrorPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const error = searchParams.get("error") || "Default";
  const errorInfo = errorMessages[error] || errorMessages.Default;

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900 p-4">
      <div className="w-full max-w-md">
        <Link
          to="/"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-6 cursor-pointer"
        >
          <ArrowLeft className="size-4" />
          返回首页
        </Link>

        <Card className="shadow-lg">
          <CardHeader className="space-y-3 text-center">
            <div className="flex justify-center mb-2">
              <div className="p-4 rounded-full bg-red-50 dark:bg-red-950/50">
                <AlertCircle className="size-8 text-red-500" />
              </div>
            </div>
            <CardTitle className="text-2xl font-bold">
              {errorInfo.title}
            </CardTitle>
            <CardDescription>
              {errorInfo.description}
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <div className="flex flex-col gap-3">
              <Button
                className="w-full h-11"
                onClick={() => navigate("/auth/signin")}
              >
                重新登录
              </Button>
              <Button
                variant="outline"
                className="w-full h-11 cursor-pointer"
                onClick={() => navigate("/")}
              >
                <Home className="size-4" data-icon="inline-start" />
                返回首页
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
