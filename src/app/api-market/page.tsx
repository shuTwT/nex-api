"use client";

import Link from "next/link";
import { useState } from "react";
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

const categories = [
  { name: "全部", icon: Grid3x3, count: 3000 },
  { name: "生活服务", icon: Sparkles, count: 856 },
  { name: "金融服务", icon: TrendingUp, count: 432 },
  { name: "人工智能", icon: Zap, count: 267 },
  { name: "天气地理", icon: Clock, count: 189 },
  { name: "企业服务", icon: Users, count: 543 },
  { name: "教育培训", icon: Star, count: 321 },
];

const popularAPIs = [
  {
    id: 1,
    name: "IP 地址查询",
    description: "县级 IP 查询定位",
    status: "normal",
    format: "json",
    method: "GET",
    price: "free",
    category: "生活服务",
    todayCalls: 1452,
    totalCalls: "237k",
    rating: 4.9,
    users: 892,
    tags: ["定位", "IP", "地理信息"]
  },
  {
    id: 2,
    name: "油价查询",
    description: "查询实时油价",
    status: "normal",
    format: "json",
    method: "GET",
    price: "premium",
    category: "生活服务",
    todayCalls: 876,
    totalCalls: "89k",
    rating: 4.7,
    users: 456,
    tags: ["油价", "能源", "实时"]
  },
  {
    id: 3,
    name: "天气查询",
    description: "全国城市天气预报",
    status: "normal",
    format: "json",
    method: "GET",
    price: "free",
    category: "天气地理",
    todayCalls: 2341,
    totalCalls: "456k",
    rating: 4.8,
    users: 1234,
    tags: ["天气", "预报", "气象"]
  },
  {
    id: 4,
    name: "短信验证",
    description: "全球短信验证码发送",
    status: "normal",
    format: "json",
    method: "POST",
    price: "premium",
    category: "企业服务",
    todayCalls: 5678,
    totalCalls: "1.2M",
    rating: 4.9,
    users: 2345,
    tags: ["短信", "验证", "通知"]
  },
  {
    id: 5,
    name: "OCR 文字识别",
    description: "支持多种语言 OCR 识别",
    status: "normal",
    format: "json",
    method: "POST",
    price: "premium",
    category: "人工智能",
    todayCalls: 1234,
    totalCalls: "345k",
    rating: 4.6,
    users: 678,
    tags: ["OCR", "识别", "AI"]
  },
  {
    id: 6,
    name: "汇率查询",
    description: "实时货币汇率查询",
    status: "normal",
    format: "json",
    method: "GET",
    price: "free",
    category: "金融服务",
    todayCalls: 987,
    totalCalls: "156k",
    rating: 4.7,
    users: 543,
    tags: ["汇率", "金融", "货币"]
  },
  {
    id: 7,
    name: "AI 绘画生成",
    description: "文本描述生成 AI 画作",
    status: "normal",
    format: "json",
    method: "POST",
    price: "premium",
    category: "人工智能",
    todayCalls: 3456,
    totalCalls: "567k",
    rating: 4.9,
    users: 1890,
    tags: ["AI", "绘画", "生成"]
  },
  {
    id: 8,
    name: "企业工商信息",
    description: "企业信息查询接口",
    status: "normal",
    format: "json",
    method: "GET",
    price: "premium",
    category: "企业服务",
    todayCalls: 765,
    totalCalls: "98k",
    rating: 4.5,
    users: 432,
    tags: ["企业", "工商", "查询"]
  },
  {
    id: 9,
    name: "快递物流查询",
    description: "支持主流快递公司",
    status: "normal",
    format: "json",
    method: "GET",
    price: "free",
    category: "生活服务",
    todayCalls: 2109,
    totalCalls: "678k",
    rating: 4.8,
    users: 1567,
    tags: ["快递", "物流", "追踪"]
  },
];

export default function ApiMarketPage() {
  const [selectedCategory, setSelectedCategory] = useState("全部");
  const [searchQuery, setSearchQuery] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [sortBy, setSortBy] = useState("popular");

  const filteredAPIs = popularAPIs.filter(api => {
    const matchCategory = selectedCategory === "全部" || api.category === selectedCategory;
    const matchSearch = api.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       api.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
                       api.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()));
    return matchCategory && matchSearch;
  });

  const sortedAPIs = [...filteredAPIs].sort((a, b) => {
    if (sortBy === "popular") return b.totalCalls.localeCompare(a.totalCalls);
    if (sortBy === "rating") return b.rating - a.rating;
    if (sortBy === "calls") return b.todayCalls - a.todayCalls;
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
                <div className="text-3xl font-bold mb-1 text-slate-900">3000+</div>
                <div className="text-sm text-slate-600">可用 API</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">80%</div>
                <div className="text-sm text-slate-600">免费接口</div>
              </div>
              <div className="bg-white/80 backdrop-blur-sm rounded-xl p-6 text-center shadow-md border border-slate-100">
                <div className="text-3xl font-bold mb-1 text-slate-900">10k+</div>
                <div className="text-sm text-slate-600">开发者</div>
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
            <div className="rounded-lg border bg-white p-4">
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
            <div className="rounded-lg border bg-white p-4">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="h-5 w-5 text-blue-600" />
                <h3 className="font-semibold">市场动态</h3>
              </div>
              <div className="space-y-3 text-sm">
                <div className="flex items-center gap-2 text-gray-600">
                  <Flame className="h-4 w-4 text-orange-500" />
                  <span>今日新增 12 个 API</span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <span>本周更新 45 个接口</span>
                </div>
                <div className="flex items-center gap-2 text-gray-600">
                  <Star className="h-4 w-4 text-yellow-500" />
                  <span>热门分类：人工智能</span>
                </div>
              </div>
            </div>
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
                    onClick={() => setSortBy("rating")}
                    className={`${sortBy === "rating" ? "bg-blue-50 border-blue-200" : ""} cursor-pointer`}
                  >
                    <Star className="h-4 w-4 mr-1" />
                    评分
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

            {/* API Grid */}
            {viewMode === "grid" ? (
              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {sortedAPIs.map((api) => (
                  <Link key={api.id} href="/api-detail">
                    <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md h-full">
                      <CardContent className="p-6 space-y-4">
                        <div className="flex items-start justify-between">
                          <div className="flex items-center gap-2">
                            <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                              正常
                            </Badge>
                            <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                              {api.format}
                            </Badge>
                          </div>
                          {api.price === "free" ? (
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
                        <div className="flex flex-wrap gap-1">
                          {api.tags.map((tag, index) => (
                            <Badge key={index} variant="secondary" className="text-xs bg-gray-100">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                        <div className="flex items-center justify-between text-xs text-gray-400">
                          <div className="flex items-center gap-1">
                            <Users className="h-3 w-3" />
                            <span>{api.users}</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <Zap className="h-3 w-3" />
                            <span>{api.todayCalls}</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <Star className="h-3 w-3 fill-current text-yellow-500" />
                            <span>{api.rating}</span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                {sortedAPIs.map((api) => (
                  <Link key={api.id} href="/api-detail">
                    <Card className="group hover:shadow-lg transition-all cursor-pointer border-0 shadow-md">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex-1 space-y-3">
                            <div className="flex items-center gap-2">
                              <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                                正常
                              </Badge>
                              <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                                {api.format}
                              </Badge>
                              <Badge variant="outline" className="bg-gray-50 text-gray-700 border-gray-200">
                                {api.method}
                              </Badge>
                              {api.price === "free" ? (
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
                            <div className="flex flex-wrap gap-2">
                              {api.tags.map((tag, index) => (
                                <Badge key={index} variant="secondary" className="text-xs bg-gray-100">
                                  {tag}
                                </Badge>
                              ))}
                            </div>
                          </div>
                          <div className="flex flex-col items-end gap-2 text-sm text-gray-500">
                            <div className="flex items-center gap-1">
                              <Users className="h-3 w-3" />
                              <span>{api.users} 开发者</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Zap className="h-3 w-3" />
                              <span>今日 {api.todayCalls} 次</span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Star className="h-3 w-3 fill-current text-yellow-500" />
                              <span>{api.rating} 分</span>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            )}

            {/* Pagination */}
            <div className="flex items-center justify-center gap-2">
              <Button variant="outline" size="sm" disabled className="cursor-pointer">上一页</Button>
              <Button variant="outline" size="sm" className="bg-blue-50 border-blue-200 cursor-pointer">1</Button>
              <Button variant="outline" size="sm" className="cursor-pointer">2</Button>
              <Button variant="outline" size="sm" className="cursor-pointer">3</Button>
              <span className="text-sm text-gray-500">...</span>
              <Button variant="outline" size="sm" className="cursor-pointer">15</Button>
              <Button variant="outline" size="sm" className="cursor-pointer">下一页</Button>
            </div>
          </main>
        </div>
      </div>
    </MainLayout>
  );
}
