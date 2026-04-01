"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Search, Star, Zap, Shield, Users, TrendingUp } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { MainLayout } from "@/components/main-layout";
import { getMarketplaceApis, getMarketplaceStats } from "@/app/actions/marketplace";
import { getPublicAnnouncement } from "@/app/actions/system-settings";

interface MarketplaceApi {
  id: string;
  name: string;
  description: string;
  alias: string;
  endpoint: string;
  method: string;
  pricing: number;
  category: string;
  isFree: boolean;
  todayCallCount: number;
  userCount: number;
  totalCallCount: number;
}

interface MarketplaceStats {
  totalApis: number;
  freeApis: number;
  paidApis: number;
  totalCallCount: number;
}

export default function Home() {
  const [apis, setApis] = useState<MarketplaceApi[]>([]);
  const [stats, setStats] = useState<MarketplaceStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [announcement, setAnnouncement] = useState<{ enabled: boolean; content: string } | null>(null);
  const [showAnnouncement, setShowAnnouncement] = useState(false);

  function checkShouldShowAnnouncement(): boolean {
    const lastDismissDate = localStorage.getItem("announcementDismissDate");
    if (!lastDismissDate) return true;
    const today = new Date().toDateString();
    return lastDismissDate !== today;
  }

  function dismissForToday() {
    localStorage.setItem("announcementDismissDate", new Date().toDateString());
    setShowAnnouncement(false);
  }

  async function loadData() {
    setIsLoading(true);
    const [apisResult, statsResult, announcementResult] = await Promise.all([
      getMarketplaceApis(),
      getMarketplaceStats(),
      getPublicAnnouncement(),
    ]);

    if (apisResult.success && apisResult.data) {
      setApis(apisResult.data);
    }

    if (statsResult.success && statsResult.data) {
      setStats(statsResult.data);
    }

    console.log(announcementResult);
    if (announcementResult.success && announcementResult.data) {
      setAnnouncement(announcementResult.data);
      if (announcementResult.data.enabled && announcementResult.data.content) {
        if (checkShouldShowAnnouncement()) {
          setShowAnnouncement(true);
        }
      }
    }

    setIsLoading(false);
  }

  useEffect(() => {
    loadData();
  }, []);

  return (
    <MainLayout>
      <Dialog open={showAnnouncement} onOpenChange={setShowAnnouncement}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="text-xl">公告</DialogTitle>
          </DialogHeader>
          <DialogDescription className="text-sm text-slate-700 whitespace-pre-wrap">
            {announcement?.content}
          </DialogDescription>
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="outline" onClick={dismissForToday}>
              今日不再提示
            </Button>
            <Button onClick={() => setShowAnnouncement(false)}>
              我知道了
            </Button>
          </div>
        </DialogContent>
      </Dialog>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-blue-50 via-white to-blue-50">
        {/* Background decorative elements */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -left-40 w-80 h-80 bg-blue-100/50 rounded-full blur-3xl"></div>
          <div className="absolute top-40 right-20 w-96 h-96 bg-blue-50/50 rounded-full blur-3xl"></div>
          <div className="absolute bottom-20 left-1/3 w-72 h-72 bg-indigo-50/50 rounded-full blur-3xl"></div>
        </div>

        <div className="container relative px-4 py-24 md:py-32 md:px-6 mx-auto">
          <div className="flex flex-col items-center text-center space-y-8">
            <div className="space-y-4">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight text-slate-900">
                发现、集成、创新
              </h1>
              <p className="text-lg md:text-xl text-slate-600 max-w-3xl mx-auto">
                一站式 API 聚合平台，为您提供<span className="font-bold text-blue-600">{stats?.totalApis || 0}+</span>高质量 API 接口，
                <span className="font-bold text-blue-600">{stats?.freeApis || 0} 免费</span>，让开发更简单
              </p>
            </div>

            {/* Search Box */}
            <div className="w-full max-w-2xl">
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
                  <Input
                    type="text"
                    placeholder="搜索 API，如：天气、短信、支付、OCR..."
                    className="pl-10 h-12 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-200 shadow-md focus:border-blue-500 focus:ring-2 focus:ring-blue-200"
                  />
                </div>
                <Button size="lg" className="h-12 px-8 bg-blue-600 hover:bg-blue-700">
                  搜索
                </Button>
              </div>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 w-full max-w-3xl mt-8">
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">{stats?.totalApis || 0}+</div>
                <div className="text-sm text-slate-600">可用 API 接口</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">{stats?.totalCallCount?.toLocaleString() || 0}+</div>
                <div className="text-sm text-slate-600">累计调用次数</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">{stats?.freeApis || 0}+</div>
                <div className="text-sm text-slate-600">免费接口</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">99.9%</div>
                <div className="text-sm text-slate-600">服务可用性</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Hot API Recommendations */}
      <section className="container px-4 py-16 md:px-6 mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h2 className="text-3xl font-bold tracking-tight mb-2">热门 API 推荐</h2>
            <p className="text-gray-500">开发者最常使用的 API 接口</p>
          </div>
          <Button variant="outline" className="gap-2 cursor-pointer">
            全部 API
            <TrendingUp className="h-4 w-4" />
          </Button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
          </div>
        ) : apis.length === 0 ? (
          <div className="text-center py-12">
            <Zap className="h-12 w-12 text-slate-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-slate-900 mb-2">暂无 API</h3>
            <p className="text-slate-500">请稍后再来查看</p>
          </div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {apis.slice(0, 5).map((api, index) => (
              <div key={api.id}>
                <Link href={`/api-detail?id=${api.id}`}>
                  <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
                    <CardContent className="p-6 space-y-4">
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                            正常
                          </Badge>
                          <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                            {api.method}
                          </Badge>
                        </div>
                        {api.isFree ? (
                          <div className="flex items-center gap-1 text-orange-500">
                            <Star className="h-4 w-4 fill-current" />
                            <span className="text-sm font-medium">免费</span>
                          </div>
                        ) : (
                          <Badge className="bg-gradient-to-r from-purple-500 to-pink-500 border-0">
                            会员专享
                          </Badge>
                        )}
                      </div>
                      <div>
                        <h3 className="text-lg font-semibold mb-1 group-hover:text-blue-600 transition-colors">
                          {api.name}
                        </h3>
                        <p className="text-sm text-gray-500">
                          {api.description}
                        </p>
                      </div>
                      <div className="flex items-center gap-4 text-xs text-gray-400">
                        <div className="flex items-center gap-1">
                          <Users className="h-3 w-3" />
                          <span>{api.userCount} 人使用</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Zap className="h-3 w-3" />
                          <span>今日：{api.todayCallCount.toLocaleString()}</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Features Section */}
      <section className="bg-gray-50 dark:bg-black py-16">
        <div className="container px-4 md:px-6 mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold tracking-tight mb-4">为什么选择我们？</h2>
            <p className="text-gray-500 max-w-2xl mx-auto">
              专业的 API 聚合平台，为开发者提供便捷、高效、稳定的 API 服务
            </p>
          </div>
          <div className="grid gap-8 md:grid-cols-3">
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="h-16 w-16 rounded-full bg-blue-100 flex items-center justify-center">
                <Zap className="h-8 w-8 text-blue-600" />
              </div>
              <h3 className="text-xl font-semibold">快速集成</h3>
              <p className="text-gray-500">
                简单的 API 接口，详细的文档说明，让您快速集成到项目中
              </p>
            </div>
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="h-16 w-16 rounded-full bg-green-100 flex items-center justify-center">
                <Shield className="h-8 w-8 text-green-600" />
              </div>
              <h3 className="text-xl font-semibold">稳定可靠</h3>
              <p className="text-gray-500">
                99.9% 的服务可用性，专业的运维团队，保障服务稳定运行
              </p>
            </div>
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="h-16 w-16 rounded-full bg-purple-100 flex items-center justify-center">
                <TrendingUp className="h-8 w-8 text-purple-600" />
              </div>
              <h3 className="text-xl font-semibold">持续更新</h3>
              <p className="text-gray-500">
                持续添加新的 API 接口，满足您不断变化的业务需求
              </p>
            </div>
          </div>
        </div>
      </section>
    </MainLayout>
  );
}
