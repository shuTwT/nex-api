"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Search,
  Plug,
  Star,
  Users,
  TrendingUp,
  Filter,
  Grid3x3,
  List,
  CheckCircle,
  Flame,
  Sparkles,
} from "lucide-react";
import { MainLayout } from "@/components/main-layout";
import { InlineAd } from "@/components/ads";
import { api } from "@/lib/api-client";
import { AdPosition } from "@/types/ad-position";

interface MarketplaceMcpService {
  id: string;
  name: string;
  identifier: string;
  type: string;
  pricing: number;
  isFree: boolean;
  todayCallCount: number;
  userCount: number;
  totalCallCount: number;
}

interface MarketplaceStats {
  totalServices: number;
  freeServices: number;
  paidServices: number;
  totalCallCount: number;
}

const TYPE_LABELS: Record<string, string> = {
  stdio: "stdio",
  sse: "SSE",
  streamableHttp: "Streamable HTTP",
};

export default function McpMarketPage() {
  const [services, setServices] = useState<MarketplaceMcpService[]>([]);
  const [stats, setStats] = useState<MarketplaceStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedType, setSelectedType] = useState("全部");
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [sortBy, setSortBy] = useState("popular");

  async function loadData() {
    setIsLoading(true);
    const [servicesResult, statsResult] = await Promise.all([
      api.get("/api/marketplace/mcp-services"),
      api.get("/api/marketplace/mcp-stats"),
    ]);

    if (servicesResult.success && servicesResult.data) {
      setServices(servicesResult.data);
    }

    if (statsResult.success && statsResult.data) {
      setStats(statsResult.data);
    }

    setIsLoading(false);
  }

  useEffect(() => {
    loadData();
  }, []);

  const types = [{ name: "全部", icon: Grid3x3, count: stats?.totalServices || 0 }];

  const uniqueTypes = [...new Set(services.map((s) => s.type))];
  uniqueTypes.forEach((t) => {
    types.push({
      name: TYPE_LABELS[t] || t,
      icon: Sparkles,
      count: services.filter((s) => s.type === t).length,
    });
  });

  const filteredServices = services.filter((s) => {
    const matchType =
      selectedType === "全部" || (TYPE_LABELS[s.type] || s.type) === selectedType;
    const matchSearch =
      s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.identifier.toLowerCase().includes(searchQuery.toLowerCase());
    return matchType && matchSearch;
  });

  const sortedServices = [...filteredServices].sort((a, b) => {
    if (sortBy === "popular") return b.totalCallCount - a.totalCallCount;
    if (sortBy === "calls") return b.todayCallCount - a.todayCallCount;
    return 0;
  });

  return (
    <MainLayout>
      <section className="relative bg-gradient-to-br from-purple-50 via-white to-purple-50">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -left-40 w-80 h-80 bg-purple-100/50 rounded-full blur-3xl" />
          <div className="absolute top-40 right-20 w-96 h-96 bg-purple-50/50 rounded-full blur-3xl" />
          <div className="absolute bottom-20 left-1/3 w-72 h-72 bg-indigo-50/50 rounded-full blur-3xl" />
        </div>

        <div className="container relative px-4 py-16 md:px-6 mx-auto">
          <div className="flex flex-col items-center text-center space-y-6">
            <div>
              <h1 className="text-4xl md:text-5xl font-bold mb-3 text-slate-900">
                MCP 市场
              </h1>
              <p className="text-lg text-slate-600">
                发现、连接 MCP 服务，为你的 AI 工具链赋能
              </p>
            </div>

            <div className="w-full max-w-2xl">
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-slate-400" />
                  <Input
                    type="text"
                    placeholder="搜索 MCP 服务名称或标识..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 h-12 bg-white text-slate-900 placeholder:text-slate-400 border border-slate-200 shadow-md focus:border-purple-500 focus:ring-2 focus:ring-purple-200"
                  />
                </div>
                <Button
                  size="lg"
                  className="h-12 px-8 bg-purple-600 hover:bg-purple-700"
                >
                  搜索
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 w-full max-w-3xl mt-8">
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">
                  {stats?.totalServices || 0}+
                </div>
                <div className="text-sm text-slate-600">可用服务</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">
                  {stats?.freeServices || 0}
                </div>
                <div className="text-sm text-slate-600">免费服务</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">
                  {stats?.totalCallCount?.toLocaleString() || 0}
                </div>
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

      <div className="flex-1 container px-4 py-8 md:px-6 mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-[240px_1fr] gap-6">
          <aside className="space-y-4">
            <div className="rounded-lg border bg-white dark:bg-black p-4">
              <div className="flex items-center gap-2 mb-4">
                <Filter className="h-5 w-5 text-purple-600" />
                <h3 className="font-semibold">服务类型</h3>
              </div>
              <div className="space-y-2">
                {types.map((type) => (
                  <button
                    key={type.name}
                    onClick={() => setSelectedType(type.name)}
                    className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-sm transition-colors cursor-pointer ${
                      selectedType === type.name
                        ? "bg-purple-50 text-purple-600"
                        : "text-gray-600 hover:bg-gray-50"
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <type.icon className="h-4 w-4" />
                      <span>{type.name}</span>
                    </div>
                    <Badge variant="outline" className="text-xs">
                      {type.count}
                    </Badge>
                  </button>
                ))}
              </div>
            </div>

            <div className="rounded-lg border bg-white dark:bg-black p-4">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="h-5 w-5 text-purple-600" />
                <h3 className="font-semibold">市场动态</h3>
              </div>
              <div className="space-y-3 text-sm">
                <div className="flex items-center gap-2 text-gray-600">
                  <Flame className="h-4 w-4 text-orange-500" />
                  <span>
                    热门类型：
                    {TYPE_LABELS[uniqueTypes[0] || ""] || uniqueTypes[0] || "暂无"}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <span>免费服务：{stats?.freeServices || 0} 个</span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <Star className="h-4 w-4 text-yellow-500" />
                  <span>付费服务：{stats?.paidServices || 0} 个</span>
                </div>
              </div>
            </div>

            <InlineAd size="lg" position={AdPosition.SIDEBAR_BOTTOM} />
          </aside>

          <main className="space-y-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="text-sm text-gray-500">
                  找到{" "}
                  <span className="font-semibold text-gray-900">
                    {sortedServices.length}
                  </span>{" "}
                  个 MCP 服务
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">排序：</span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSortBy("popular")}
                    className={`${sortBy === "popular" ? "bg-purple-50 border-purple-200" : ""} cursor-pointer`}
                  >
                    <Flame className="h-4 w-4 mr-1" />
                    热门
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSortBy("calls")}
                    className={`${sortBy === "calls" ? "bg-purple-50 border-purple-200" : ""} cursor-pointer`}
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
                  className={`${viewMode === "grid" ? "bg-purple-50" : ""} cursor-pointer`}
                >
                  <Grid3x3 className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setViewMode("list")}
                  className={`${viewMode === "list" ? "bg-purple-50" : ""} cursor-pointer`}
                >
                  <List className="h-4 w-4" />
                </Button>
              </div>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent" />
              </div>
            ) : sortedServices.length === 0 ? (
              <div className="text-center py-12">
                <Plug className="h-12 w-12 text-slate-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-slate-900 mb-2">
                  暂无 MCP 服务
                </h3>
                <p className="text-slate-500">请稍后再来查看</p>
              </div>
            ) : viewMode === "grid" ? (
              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {sortedServices.map((svc) => (
                  <div key={svc.id}>
                    <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
                      <CardContent className="p-6 space-y-4">
                        <div className="flex items-start justify-between">
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className="bg-green-50 text-green-700 border-green-200"
                            >
                              正常
                            </Badge>
                            <Badge
                              variant="outline"
                              className="bg-purple-50 text-purple-700 border-purple-200"
                            >
                              {TYPE_LABELS[svc.type] || svc.type}
                            </Badge>
                          </div>
                          {svc.isFree ? (
                            <div className="flex items-center gap-1 text-orange-500">
                              <Star className="h-4 w-4 fill-current" />
                              <span className="text-sm font-medium">免费</span>
                            </div>
                          ) : (
                            <Badge className="bg-gradient-to-r from-purple-500 to-pink-500 border-0">
                              付费
                            </Badge>
                          )}
                        </div>
                        <div>
                          <h3 className="text-lg font-semibold mb-1 group-hover:text-purple-600 transition-colors">
                            {svc.name}
                          </h3>
                          <code className="text-xs text-purple-500 bg-purple-50 px-2 py-0.5 rounded">
                            /api/v1/mcp/{svc.identifier}
                          </code>
                        </div>
                        <div className="flex items-center justify-between text-xs text-gray-400">
                          <div className="flex items-center gap-1">
                            <Users className="h-3 w-3" />
                            <span>{svc.userCount} 人使用</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <Plug className="h-3 w-3" />
                            <span>今日：{svc.todayCallCount.toLocaleString()}</span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                {sortedServices.map((svc) => (
                  <Card
                    key={svc.id}
                    className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md"
                  >
                    <CardContent className="p-6">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 space-y-3">
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className="bg-green-50 text-green-700 border-green-200"
                            >
                              正常
                            </Badge>
                            <Badge
                              variant="outline"
                              className="bg-purple-50 text-purple-700 border-purple-200"
                            >
                              {TYPE_LABELS[svc.type] || svc.type}
                            </Badge>
                            {svc.isFree ? (
                              <div className="flex items-center gap-1 text-orange-500">
                                <Star className="h-4 w-4 fill-current" />
                                <span className="text-sm font-medium">免费</span>
                              </div>
                            ) : (
                              <Badge className="bg-gradient-to-r from-purple-500 to-pink-500 border-0">
                                付费
                              </Badge>
                            )}
                          </div>
                          <div>
                            <h3 className="text-lg font-semibold mb-1 group-hover:text-purple-600 transition-colors">
                              {svc.name}
                            </h3>
                            <code className="text-xs text-purple-500 bg-purple-50 px-2 py-0.5 rounded">
                              /api/v1/mcp/{svc.identifier}
                            </code>
                          </div>
                        </div>
                        <div className="flex flex-col items-end gap-2 text-sm text-gray-500">
                          <div className="flex items-center gap-1">
                            <Users className="h-3 w-3" />
                            <span>{svc.userCount} 人使用</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <Plug className="h-3 w-3" />
                            <span>今日 {svc.todayCallCount.toLocaleString()} 次</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <TrendingUp className="h-3 w-3" />
                            <span>累计 {svc.totalCallCount.toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </main>
        </div>
      </div>
    </MainLayout>
  );
}
