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
import { Github, MessageCircle, Shield } from "lucide-react";

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuthModal({ open, onOpenChange }: AuthModalProps) {
  const [isLoading, setIsLoading] = useState<string | null>(null);

  const handleOAuthLogin = async (provider: string) => {
    setIsLoading(provider);
    try {
      window.location.href = `/api/auth/${provider}`;
    } catch (error) {
      console.error(`${provider} login error:`, error);
      setIsLoading(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader className="space-y-3">
          <DialogTitle className="text-2xl font-bold text-center">
            欢迎来到 One API
          </DialogTitle>
          <DialogDescription className="text-center text-base">
            使用第三方账号快速登录或注册
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3 mt-6">
          <Button
            variant="outline"
            size="lg"
            className="w-full h-12 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-gray-100 hover:border-gray-400"
            onClick={() => handleOAuthLogin("github")}
            disabled={isLoading !== null}
          >
            {isLoading === "github" ? (
              <div className="flex items-center gap-2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-gray-600 border-t-transparent" />
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
            className="w-full h-12 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-indigo-50 hover:border-indigo-400 hover:text-indigo-600"
            onClick={() => handleOAuthLogin("discord")}
            disabled={isLoading !== null}
          >
            {isLoading === "discord" ? (
              <div className="flex items-center gap-2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent" />
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
            className="w-full h-12 text-base font-medium cursor-pointer transition-all duration-200 hover:bg-purple-50 hover:border-purple-400 hover:text-purple-600"
            onClick={() => handleOAuthLogin("sso")}
            disabled={isLoading !== null}
          >
            {isLoading === "sso" ? (
              <div className="flex items-center gap-2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-purple-600 border-t-transparent" />
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

        <div className="mt-6 pt-6 border-t border-gray-200">
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
