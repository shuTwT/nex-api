import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LogOut, ArrowLeft } from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { useNavigate, Link } from "react-router";

export default function AuthSignoutPage() {
  const navigate = useNavigate();
  const { logout } = useAuth();

  const handleSignOut = async () => {
    document.cookie = "nex-api-auth=; Max-Age=0; path=/; SameSite=Lax";
    await logout();
    navigate("/", { replace: true });
  };

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
              <div className="p-4 rounded-full bg-muted">
                <LogOut className="size-8 text-muted-foreground" />
              </div>
            </div>
            <CardTitle className="text-2xl font-bold">
              退出登录
            </CardTitle>
            <CardDescription>
              确定要退出您的账户吗？
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <div className="flex flex-col gap-3">
              <Button
                className="w-full h-11"
                onClick={handleSignOut}
              >
                <LogOut className="size-4" data-icon="inline-start" />
                确认退出
              </Button>
              <Button
                variant="outline"
                className="w-full h-11 cursor-pointer"
                onClick={() => navigate(-1)}
              >
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
