"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { initializeSystem } from "@/app/actions/initialize";
import { Zap, Shield, Users, Sparkles, CheckCircle2 } from "lucide-react";

export function InitializationForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setError(null);

    try {
      const result = await initializeSystem(formData);

      if (result.success) {
        window.location.href = "/console";
      } else {
        setError(result.error || "初始化失败");
      }
    } catch (err) {
      setError("初始化失败，请重试");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-5xl grid md:grid-cols-2 gap-8 items-center">
        {/* Left Side - Branding */}
        <div className="space-y-8">
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-blue-600 to-cyan-600 flex items-center justify-center shadow-lg">
                <Zap className="h-7 w-7 text-white" />
              </div>
              <div>
                <h1 className="text-3xl font-bold text-slate-900">NexApi</h1>
                <p className="text-sm text-slate-500">聚合管理系统</p>
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <h2 className="text-4xl font-bold text-slate-900 leading-tight">
              欢迎使用<br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-cyan-600">
                NexApi
              </span>
            </h2>
            <p className="text-lg text-slate-600 leading-relaxed">
              开始您的 API 管理之旅。创建第一个管理员账户，<br />
              即可开始使用系统的全部功能。
            </p>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="bg-white/60 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
              <Shield className="h-8 w-8 text-blue-600 mb-2" />
              <p className="text-sm font-medium text-slate-900">安全可靠</p>
              <p className="text-xs text-slate-500 mt-1">企业级安全防护</p>
            </div>
            <div className="bg-white/60 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
              <Users className="h-8 w-8 text-cyan-600 mb-2" />
              <p className="text-sm font-medium text-slate-900">团队协作</p>
              <p className="text-xs text-slate-500 mt-1">多用户权限管理</p>
            </div>
            <div className="bg-white/60 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
              <Sparkles className="h-8 w-8 text-purple-600 mb-2" />
              <p className="text-sm font-medium text-slate-900">智能管理</p>
              <p className="text-xs text-slate-500 mt-1">一站式 API 管理</p>
            </div>
          </div>
        </div>

        {/* Right Side - Form */}
        <div className="bg-white rounded-2xl shadow-xl border border-slate-200 p-8">
          <div className="space-y-6">
            <div className="text-center space-y-2">
              <div className="inline-flex items-center gap-2 bg-blue-50 text-blue-700 px-3 py-1 rounded-full text-sm font-medium">
                <Sparkles className="h-4 w-4" />
                系统初始化
              </div>
              <h3 className="text-2xl font-bold text-slate-900">创建管理员账户</h3>
              <p className="text-slate-500 text-sm">这是系统的第一个账户，将自动成为管理员</p>
            </div>

            <form action={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">
                  邮箱地址 <span className="text-red-500">*</span>
                </label>
                <Input
                  type="email"
                  name="email"
                  placeholder="admin@example.com"
                  required
                  disabled={isLoading}
                  className="h-11"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">
                  用户名 <span className="text-red-500">*</span>
                </label>
                <Input
                  type="text"
                  name="username"
                  placeholder="admin"
                  required
                  disabled={isLoading}
                  className="h-11"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">
                  密码 <span className="text-red-500">*</span>
                </label>
                <Input
                  type="password"
                  name="password"
                  placeholder="至少 8 位字符"
                  required
                  disabled={isLoading}
                  className="h-11"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-700">
                  确认密码 <span className="text-red-500">*</span>
                </label>
                <Input
                  type="password"
                  name="confirmPassword"
                  placeholder="再次输入密码"
                  required
                  disabled={isLoading}
                  className="h-11"
                />
              </div>

              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              )}

              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-blue-900">首次使用提示</p>
                    <p className="text-xs text-blue-700">
                      此账户将获得管理员权限，可以管理所有用户和 API 接口。
                      请妥善保管您的登录凭证。
                    </p>
                  </div>
                </div>
              </div>

              <Button
                type="submit"
                disabled={isLoading}
                className="w-full h-11 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 cursor-pointer"
              >
                {isLoading ? "初始化中..." : "完成初始化"}
              </Button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
