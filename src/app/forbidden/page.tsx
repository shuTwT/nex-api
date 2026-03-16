"use client";

import { MainLayout } from "@/components/main-layout";
import { Button } from "@/components/ui/button";
import { ShieldX, ArrowLeft } from "lucide-react";
import Link from "next/link";

export default function ForbiddenPage() {
  return (
    <MainLayout>
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-center px-4">
          <div className="mb-8">
            <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-orange-100 mb-6">
              <ShieldX className="h-12 w-12 text-orange-600" />
            </div>
          </div>
          
          <h1 className="text-7xl font-bold text-slate-900 mb-4">
            403
          </h1>
          
          <h2 className="text-2xl font-semibold text-slate-700 mb-4">
            禁止访问
          </h2>
          
          <p className="text-lg text-slate-500 mb-8 max-w-md mx-auto">
            您没有权限访问此页面。此页面仅对管理员开放。
          </p>
          
          <div className="flex gap-4 justify-center">
            <Link href="/">
              <Button variant="outline" className="gap-2">
                <ArrowLeft className="h-4 w-4" />
                返回首页
              </Button>
            </Link>
            <Link href="/console">
              <Button className="bg-blue-600 hover:bg-blue-700">
                前往控制台
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
