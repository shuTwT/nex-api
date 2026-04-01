"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Mail, ArrowLeft, Home } from "lucide-react";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function VerifyRequestPage() {
  const router = useRouter();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900 p-4">
      <div className="w-full max-w-md">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-6 cursor-pointer"
        >
          <ArrowLeft className="size-4" />
          返回首页
        </Link>

        <Card className="shadow-lg">
          <CardHeader className="space-y-3 text-center">
            <div className="flex justify-center mb-2">
              <div className="p-4 rounded-full bg-green-50 dark:bg-green-950/50">
                <Mail className="size-8 text-green-500" />
              </div>
            </div>
            <CardTitle className="text-2xl font-bold">
              请检查您的邮箱
            </CardTitle>
            <CardDescription>
              我们已向您发送了验证链接，请点击链接完成登录。
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <div className="p-4 rounded-lg bg-muted">
              <p className="text-sm text-muted-foreground text-center">
                如果您没有收到邮件，请检查垃圾邮件文件夹，或稍后重试。
              </p>
            </div>

            <div className="flex flex-col gap-3">
              <Button
                variant="outline"
                className="w-full h-11 cursor-pointer"
                onClick={() => router.push("/")}
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
