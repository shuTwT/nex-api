"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Search, Star, Zap, Shield, Users, TrendingUp } from "lucide-react";
import { MainLayout } from "@/components/main-layout";

export default function Home() {
  return (
    <MainLayout>
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
                一站式 API 聚合平台，为您提供<span className="font-bold text-blue-600">3000+</span>高质量 API 接口，
                <span className="font-bold text-blue-600">80% 免费</span>，让开发更简单
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

            {/* Hot Tags */}
            <div className="flex flex-wrap items-center justify-center gap-2 text-sm">
              <span className="text-slate-600">热门搜索：</span>
              <Badge variant="secondary" className="bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 cursor-pointer">天气 API</Badge>
              <Badge variant="secondary" className="bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 cursor-pointer">短信验证</Badge>
              <Badge variant="secondary" className="bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 cursor-pointer">IP 查询</Badge>
              <Badge variant="secondary" className="bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 cursor-pointer">OCR 识别</Badge>
              <Badge variant="secondary" className="bg-blue-50 hover:bg-blue-100 text-blue-700 border border-blue-200 cursor-pointer">AI 绘画</Badge>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 w-full max-w-3xl mt-8">
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">2+</div>
                <div className="text-sm text-slate-600">可用 API 接口</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">237+</div>
                <div className="text-sm text-slate-600">累计调用次数</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">1+</div>
                <div className="text-sm text-slate-600">注册用户</div>
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

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* API Card 1 */}
          <Link href="/api-detail">
            <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
              <CardContent className="p-6 space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                      正常
                    </Badge>
                    <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                      json
                    </Badge>
                  </div>
                  <div className="flex items-center gap-1 text-orange-500">
                    <Star className="h-4 w-4 fill-current" />
                    <span className="text-sm font-medium">免费</span>
                  </div>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-1 group-hover:text-blue-600 transition-colors">
                    IP 地址查询
                  </h3>
                  <p className="text-sm text-gray-500">
                    县级 IP 查询定位
                  </p>
                </div>
                <div className="flex items-center gap-4 text-xs text-gray-400">
                  <div className="flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    <span>今日调用：1,452</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Zap className="h-3 w-3" />
                    <span>累计调用：237k</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </Link>

          {/* API Card 2 */}
          <Link href="/api-detail">
            <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
              <CardContent className="p-6 space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                      正常
                    </Badge>
                    <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                      json
                    </Badge>
                  </div>
                  <Badge className="bg-gradient-to-r from-purple-500 to-pink-500 border-0">
                    会员专享
                  </Badge>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-1 group-hover:text-blue-600 transition-colors">
                    油价查询
                  </h3>
                  <p className="text-sm text-gray-500">
                    查询实时油价
                  </p>
                </div>
                <div className="flex items-center gap-4 text-xs text-gray-400">
                  <div className="flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    <span>今日调用：876</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Zap className="h-3 w-3" />
                    <span>累计调用：89k</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </Link>

          {/* API Card 3 */}
          <Link href="/api-detail">
            <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
              <CardContent className="p-6 space-y-4">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                      正常
                    </Badge>
                    <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                      json
                    </Badge>
                  </div>
                  <div className="flex items-center gap-1 text-orange-500">
                    <Star className="h-4 w-4 fill-current" />
                    <span className="text-sm font-medium">免费</span>
                  </div>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-1 group-hover:text-blue-600 transition-colors">
                    天气查询
                  </h3>
                  <p className="text-sm text-gray-500">
                    全国城市天气预报
                  </p>
                </div>
                <div className="flex items-center gap-4 text-xs text-gray-400">
                  <div className="flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    <span>今日调用：2,341</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Zap className="h-3 w-3" />
                    <span>累计调用：456k</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </Link>
        </div>
      </section>

      {/* Features Section */}
      <section className="bg-gray-50 py-16">
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
