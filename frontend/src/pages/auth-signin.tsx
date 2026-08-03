import { useState, type FormEvent } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Github, Mail, Shield, Loader2, ArrowLeft } from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { useOAuthProviders } from "@/hooks/use-oauth-providers";
import { useNavigate, useSearchParams, Link } from "react-router";

export default function AuthSigninPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const callbackUrl = searchParams.get("callbackUrl") || "/";
  const { login } = useAuth();
  const { providers, isLoading: isLoadingProviders } = useOAuthProviders();

  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleOAuthLogin = (provider: string) => {
    setIsLoading(provider);
    window.location.href = `/api/auth/signin/${provider}?callbackUrl=${encodeURIComponent(callbackUrl)}`;
  };

  const handleCredentialsLogin = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading("credentials");
    try {
      const success = await login(email, password);
      if (!success) {
        setError("邮箱或密码错误");
        setIsLoading(null);
        return;
      }
      navigate(callbackUrl, { replace: true });
    } catch {
      setError("登录失败，请稍后重试");
      setIsLoading(null);
    }
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
          <CardHeader className="space-y-3">
            <CardTitle className="text-2xl font-bold text-center">
              欢迎回来
            </CardTitle>
            <CardDescription className="text-center">
              登录您的 NexApi 账户
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <form onSubmit={handleCredentialsLogin} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">邮箱</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="请输入邮箱"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isLoading !== null}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">密码</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="请输入密码"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isLoading !== null}
                />
              </div>

              {error && (
                <p className="text-sm text-red-500 text-center">{error}</p>
              )}

              <Button
                type="submit"
                className="w-full h-11"
                disabled={isLoading !== null}
              >
                {isLoading === "credentials" ? (
                  <div className="flex items-center gap-2">
                    <Loader2 className="size-4 animate-spin" />
                    <span>登录中...</span>
                  </div>
                ) : (
                  <>
                    <Mail className="size-4" data-icon="inline-start" />
                    邮箱登录
                  </>
                )}
              </Button>
            </form>

            {!isLoadingProviders && providers.length > 0 && (
              <>
                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <Separator />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-card px-2 text-muted-foreground">
                      或使用第三方账号
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-3">
                  {providers.map((provider) => {
                    const Icon = provider.id === "github" ? Github : Shield;
                    return (
                      <Button
                        key={provider.id}
                        variant="outline"
                        className="w-full h-11 cursor-pointer transition-all duration-200 hover:bg-accent"
                        onClick={() => handleOAuthLogin(provider.id)}
                        disabled={isLoading !== null}
                      >
                        {isLoading === provider.id ? (
                          <div className="flex items-center gap-2">
                            <Loader2 className="size-4 animate-spin" />
                            <span>正在跳转...</span>
                          </div>
                        ) : (
                          <>
                            <Icon className="size-4" data-icon="inline-start" />
                            使用 {provider.name} 登录
                          </>
                        )}
                      </Button>
                    );
                  })}
                </div>
              </>
            )}

            <div className="pt-4 border-t">
              <p className="text-xs text-center text-muted-foreground leading-relaxed">
                登录即表示您同意我们的
                <Link
                  to="/terms"
                  className="text-primary hover:underline cursor-pointer mx-1"
                >
                  服务条款
                </Link>
                和
                <Link
                  to="/privacy"
                  className="text-primary hover:underline cursor-pointer mx-1"
                >
                  隐私政策
                </Link>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
