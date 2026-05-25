"use client";

import { useState } from "react";
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
import { Github, MessageCircle, Shield, Mail, Loader2 } from "lucide-react";
import { signIn } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuthModal({ open, onOpenChange }: AuthModalProps) {
  const { fetchUserInfo } = useAuthStore();
  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const router = useRouter();

  const handleOAuthLogin = async (provider: string) => {
    setIsLoading(provider);
    try {
      await signIn(provider);
    } catch (error) {
      console.error(`${provider} login error:`, error);
      setIsLoading(null);
    }
  };

  const handleCredentialsLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading("credentials");
    try {
      const result = await signIn("credentials", {
        email,
        password,
        redirect: false,
      });

      console.log("result",result);
      if (result?.error) {
        setError("邮箱或密码错误");
        setIsLoading(null);
        return;
      }

      onOpenChange(false);
      await fetchUserInfo();
      router.refresh();
    } catch (error) {
      console.error("Credentials login error:", error);
      setError("登录失败，请稍后重试");
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

        {/* Email/Password Login Form */}
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
                <span>登录中...</span>
              </div>
            ) : (
              <>
                <Mail className="mr-2 h-4 w-4" />
                邮箱登录
              </>
            )}
          </Button>
        </form>

        {/* Divider */}
        <div className="relative my-4">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t border-gray-200" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-white px-2 text-gray-500">或使用第三方账号</span>
          </div>
        </div>

        {/* OAuth Providers */}
        <div className="flex flex-col gap-3">
          <Button
            variant="outline"
            size="lg"
            className="w-full h-11 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-gray-100 hover:border-gray-400"
            onClick={() => handleOAuthLogin("github")}
            disabled={isLoading !== null}
          >
            {isLoading === "github" ? (
              <div className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>正在跳转...</span>
              </div>
            ) : (
              <>
                <Github className="mr-2 h-5 w-5" />
                使用 GitHub 登录
              </>
            )}
          </Button>

          <Button
            variant="outline"
            size="lg"
            className="w-full h-11 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-indigo-50 hover:border-indigo-400 hover:text-indigo-600"
            onClick={() => handleOAuthLogin("discord")}
            disabled={isLoading !== null}
          >
            {isLoading === "discord" ? (
              <div className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>正在跳转...</span>
              </div>
            ) : (
              <>
                <MessageCircle className="mr-2 h-5 w-5" />
                使用 Discord 登录
              </>
            )}
          </Button>

          <Button
            variant="outline"
            size="lg"
            className="w-full h-11 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-purple-50 hover:border-purple-400 hover:text-purple-600"
            onClick={() => handleOAuthLogin("easy1auth")}
            disabled={isLoading !== null}
          >
            {isLoading === "easy1auth" ? (
              <div className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>正在跳转...</span>
              </div>
            ) : (
              <>
                <Shield className="mr-2 h-5 w-5" />
                使用统一身份认证
              </>
            )}
          </Button>
        </div>

        <div className="mt-6 pt-4 border-t border-gray-200">
          <p className="text-xs text-center text-gray-500 leading-relaxed">
            登录即表示您同意我们的
            <a
              href="/terms"
              className="text-blue-600 hover:text-blue-700 underline cursor-pointer mx-1"
            >
              服务条款
            </a>
            和
            <a
              href="/privacy"
              className="text-blue-600 hover:text-blue-700 underline cursor-pointer mx-1"
            >
              隐私政策
            </a>
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}
