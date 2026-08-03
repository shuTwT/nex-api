import { useState, type FormEvent } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Github, Shield } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { useAuth } from "@/hooks/use-auth";
import { useOAuthProviders } from "@/hooks/use-oauth-providers";

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuthModal({ open, onOpenChange }: AuthModalProps) {
  const { login } = useAuth();
  const { providers, isLoading: isLoadingProviders } = useOAuthProviders();
  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleOAuthLogin = (provider: string) => {
    setIsLoading(provider);
    // 弹窗无 callbackUrl 参数，回跳当前页面。
    const callbackUrl = window.location.pathname + window.location.search;
    window.location.href = `/api/auth/signin/${provider}?callbackUrl=${encodeURIComponent(callbackUrl)}`;
  };

  const handleCredentialsLogin = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading("credentials");
    try {
      const success = await login(email, password);
      if (success) {
        onOpenChange(false);
      } else {
        setError("邮箱或密码错误");
      }
    } catch {
      setError("登录失败，请稍后重试");
    } finally {
      setIsLoading(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader className="space-y-3">
          <DialogTitle className="text-2xl font-bold text-center">
            欢迎来到 NexApi
          </DialogTitle>
          <DialogDescription className="text-center text-base">
            登录或注册您的账户
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleCredentialsLogin} className="space-y-4 mt-4">
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
                <Loader2 className="h-4 w-4 animate-spin" />
                登录中...
              </div>
            ) : (
              "登录"
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
                    type="button"
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
      </DialogContent>
    </Dialog>
  );
}
