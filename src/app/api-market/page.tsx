"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { 
  Search, 
  Zap, 
  Star, 
  Users, 
  TrendingUp, 
  Filter, 
  Grid3x3, 
  List,
  CheckCircle,
  Clock,
  Flame,
  Sparkles
} from "lucide-react";
import { MainLayout } from "@/components/main-layout";
import { InlineAd } from "@/components/ads";
import { getMarketplaceApis, getMarketplaceStats } from "@/app/actions/marketplace";
import { AdPosition } from "@/types/ad-position";

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

export default function ApiMarketPage() {
  const [apis, setApis] = useState<MarketplaceApi[]>([]);
  const [stats, setStats] = useState<MarketplaceStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState("全部");
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [sortBy, setSortBy] = useState("popular");

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setIsLoading(true);
    const [apisResult, statsResult] = await Promise.all([
      getMarketplaceApis(),
      getMarketplaceStats(),
    ]);

    if (apisResult.success && apisResult.data) {
      setApis(apisResult.data);
    }

    if (statsResult.success && statsResult.data) {
      setStats(statsResult.data);
    }

    setIsLoading(false);
  }

  const categories = [
    { name: "全部", icon: Grid3x3, count: stats?.totalApis || 0 },
  ];

  const uniqueCategories = [...new Set(apis.map(api => api.category))];
  uniqueCategories.forEach(cat => {
    categories.push({
      name: cat,
      icon: Sparkles,
      count: apis.filter(api => api.category === cat).length,
    });
  });

  const filteredAPIs = apis.filter(api => {
    const matchCategory = selectedCategory === "全部" || api.category === selectedCategory;
    const matchSearch = api.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       api.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchCategory && matchSearch;
  });

  const sortedAPIs = [...filteredAPIs].sort((a, b) => {
    if (sortBy === "popular") return b.totalCallCount - a.totalCallCount;
    if (sortBy === "calls") return b.todayCallCount - a.todayCallCount;
    return 0;
  });

  return (
    <MainLayout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-blue-50 via-white to-blue-50">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -left-40 w-80 h-80 bg-blue-100/50 rounded-full blur-3xl"></div>
          <div className="absolute top-40 right-20 w-96 h-96 bg-blue-50/50 rounded-full blur-3xl"></div>
          <div className="absolute bottom-20 left-1/3 w-72 h-72 bg-indigo-50/50 rounded-full blur-3xl"></div>
        </div>

        <div className="container relative px-4 py-16 md:px-6 mx-auto">
          <div className="flex flex-col items-center text-center space-y-6">
            <div>
              <h1 className="text-4xl md:text-5xl font-bold mb-3 text-slate-900">
                API 市场
              </h1>
              <p className="text-lg text-slate-600">
                发现、探索、集成优质 API 接口
              </p>
            </div>

            {/* Search Bar */}
            <div className="w-full max-w-2xl">
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
                  <Input
                    type="text"
                    placeholder="搜索 API、分类或标签..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
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
                <div className="text-sm text-slate-600">可用 API</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">{stats?.freeApis || 0}</div>
                <div className="text-sm text-slate-600">免费接口</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">{stats?.totalCallCount?.toLocaleString() || 0}</div>
                <div className="text-sm text-slate-600">累计调用</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">99.9%</div>
                <div className="text-sm text-slate-600">服务可用性</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Main Content */}
      <div className="flex-1 container px-4 py-8 md:px-6 mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-[240px_1fr] gap-6">
          {/* Sidebar - Categories */}
          <aside className="space-y-4">
            <div className="rounded-lg border bg-white dark:bg-black p-4">
              <div className="flex items-center gap-2 mb-4">
                <Filter className="h-5 w-5 text-blue-600" />
                <h3 className="font-semibold">API 分类</h3>
              </div>
              <div className="space-y-2">
                {categories.map((category) => (
                  <button
                    key={category.name}
                    onClick={() => setSelectedCategory(category.name)}
                    className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-sm transition-colors cursor-pointer ${
                      selectedCategory === category.name
                        ? "bg-blue-50 text-blue-600"
                        : "text-gray-600 hover:bg-gray-50"
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <category.icon className="h-4 w-4" />
                      <span>{category.name}</span>
                    </div>
                    <Badge variant="outline" className="text-xs">
                      {category.count}
                    </Badge>
                  </button>
                ))}
              </div>
            </div>

            {/* Quick Stats */}
            <div className="rounded-lg border bg-white dark:bg-black p-4">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="h-5 w-5 text-blue-600" />
                <h3 className="font-semibold">市场动态</h3>
              </div>
              <div className="space-y-3 text-sm">
                <div className="flex items-center gap-2 text-gray-600">
                  <Flame className="h-4 w-4 text-orange-500" />
                  <span>热门分类：{uniqueCategories[0] || "暂无"}</span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <span>免费接口：{stats?.freeApis || 0} 个</span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <Star className="h-4 w-4 text-yellow-500" />
                  <span>会员专享：{stats?.paidApis || 0} 个</span>
                </div>
              </div>
            </div>

            {/* Sidebar Ad */}
            <InlineAd size="lg" position={AdPosition.SIDEBAR_BOTTOM} />
          </aside>

          {/* Main Content */}
          <main className="space-y-6">
            {/* Toolbar */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="text-sm text-gray-500">
                  找到 <span className="font-semibold text-gray-900">{sortedAPIs.length}</span> 个 API
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">排序：</span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSortBy("popular")}
                    className={`${sortBy === "popular" ? "bg-blue-50 border-blue-200" : ""} cursor-pointer`}
                  >
                    <Flame className="h-4 w-4 mr-1" />
                    热门
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSortBy("calls")}
                    className={`${sortBy === "calls" ? "bg-blue-50 border-blue-200" : ""} cursor-pointer`}
                  >
                    <TrendingUp className="h-4 w-4 mr-1" />
                    调用量
                  </Button>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setViewMode("grid")}
                  className={`${viewMode === "grid" ? "bg-blue-50" : ""} cursor-pointer`}
                >
                  <Grid3x3 className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setViewMode("list")}
                  className={`${viewMode === "list" ? "bg-blue-50" : ""} cursor-pointer`}
                >
                  <List className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {/* Loading State */}
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
              </div>
            ) : sortedAPIs.length === 0 ? (
              <div className="text-center py-12">
                <Zap className="h-12 w-12 text-slate-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-slate-900 mb-2">暂无 API</h3>
                <p className="text-slate-500">请稍后再来查看</p>
              </div>
            ) : viewMode === "grid" ? (
              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {sortedAPIs.map((api, index) => (
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
                          <div className="flex items-center justify-between text-xs text-gray-400">
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
            ) : (
              <div className="space-y-4">
                {sortedAPIs.map((api) => (
                  <Link key={api.id} href={`/api-detail?id=${api.id}`}>
                    <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex-1 space-y-3">
                            <div className="flex items-center gap-2">
                              <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                                正常
                              </Badge>
                              <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                                {api.method}
                              </Badge>
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
                          </div>
                          <div className="flex flex-col items-end gap-2 text-sm text-gray-500">
                            <div className="flex items-center gap-1">
                              <Users className="h-3 w-3" />
                              <span>{api.userCount} 人使用</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Zap className="h-3 w-3" />
                              <span>今日 {api.todayCallCount.toLocaleString()} 次</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <TrendingUp className="h-3 w-3" />
                              <span>累计 {api.totalCallCount.toLocaleString()}</span>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            )}
          </main>
        </div>
      </div>
    </MainLayout>
  );
}
