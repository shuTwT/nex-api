"use client";

import { useEffect, useState, useTransition } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Search,
  Plus,
  Edit,
  Trash2,
  Settings,
  Zap,
  Activity,
  Pause,
  Database,
  FolderOpen,
} from "lucide-react";
import {
  getApis,
  getApiStats,
  toggleApiStatus,
  deleteApi,
} from "@/app/actions/apis";
import { getCategories } from "@/app/actions/categories";
import { Pagination } from "@/components/pagination";
import { ApiForm } from "@/components/api-form";
import { CategoryForm } from "@/components/category-form";

export interface Api {
  id: string;
  name: string;
  alias: string;
  description: string;
  endpoint: string;
  method: string;
  category: {
    id: string;
    name: string;
  };
  pricing: string;
  documentation: string | null;
  preScript: string | null;
  postScript: string | null;
  isActive: boolean;
  callCount: number;
  createdAt: string;
}

interface Category {
  id: string;
  name: string;
  description: string;
  icon: string | null;
  apiCount: number;
}

interface ApiStats {
  totalApis: number;
  activeApis: number;
  inactiveApis: number;
  totalCalls: number;
  categoriesCount: number;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export default function APIManagementPage() {
  const [apis, setApis] = useState<Api[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [stats, setStats] = useState<ApiStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [selectedStatus, setSelectedStatus] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [isPending, startTransition] = useTransition();
  const [showApiForm, setShowApiForm] = useState(false);
  const [editingApi, setEditingApi] = useState<Api | null>(null);
  const [showCategoryForm, setShowCategoryForm] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);

  useEffect(() => {
    loadApis();
    loadCategories();
    loadStats();
  }, [searchQuery, selectedCategory, selectedStatus, currentPage, pageSize]);

  async function loadApis() {
    setIsLoading(true);
    const result = await getApis({
      category: selectedCategory,
      search: searchQuery,
      status: selectedStatus,
      page: currentPage,
      limit: pageSize,
    });

    if (result.success && result.data) {
      setApis(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }

  async function loadCategories() {
    const result = await getCategories();
    if (result.success && result.data) {
      setCategories(result.data);
    }
  }

  async function loadStats() {
    const result = await getApiStats();
    if (result.success && result.data) {
      setStats(result.data);
    }
  }

  function handleSearchChange(value: string) {
    setSearchQuery(value);
    setCurrentPage(1);
  }

  function handleCategoryChange(category: string) {
    setSelectedCategory(category);
    setCurrentPage(1);
  }

  function handleStatusChange(status: string) {
    setSelectedStatus(status);
    setCurrentPage(1);
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  async function handleToggleStatus(id: string) {
    startTransition(async () => {
      const result = await toggleApiStatus(id);
      if (result.success) {
        loadApis();
        loadStats();
      } else {
        alert(result.error || "切换状态失败");
      }
    });
  }

  async function handleDelete(id: string) {
    if (!confirm("确定要删除这个 API 吗？此操作无法撤销。")) {
      return;
    }

    startTransition(async () => {
      const result = await deleteApi(id);
      if (result.success) {
        loadApis();
        loadStats();
      } else {
        alert(result.error || "删除失败");
      }
    });
  }

  function handleAddApi() {
    setEditingApi(null);
    setShowApiForm(true);
  }

  function handleEditApi(api: Api) {
    setEditingApi(api);
    setShowApiForm(true);
  }

  function handleApiFormSuccess() {
    startTransition(() => {
      loadApis();
      loadCategories();
      loadStats();
    });
  }

  function handleAddCategory() {
    setEditingCategory(null);
    setShowCategoryForm(true);
  }

  function handleEditCategory(category: Category) {
    setEditingCategory(category);
    setShowCategoryForm(true);
  }

  function handleCategoryFormSuccess() {
    startTransition(() => {
      loadCategories();
      loadApis();
    });
  }

  const statsCards = [
    {
      title: "总接口数",
      value: stats?.totalApis || 0,
      icon: Zap,
      color: "blue",
    },
    {
      title: "活跃接口",
      value: stats?.activeApis || 0,
      icon: Activity,
      color: "green",
    },
    {
      title: "停用接口",
      value: stats?.inactiveApis || 0,
      icon: Pause,
      color: "orange",
    },
    {
      title: "总调用次数",
      value: (stats?.totalCalls || 0).toLocaleString(),
      icon: Database,
      color: "purple",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">接口管理</h1>
          <p className="text-slate-500 mt-1">管理系统 API 接口</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleAddApi}>
          <Plus className="h-4 w-4" />
          添加接口
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            green: "bg-green-50 text-green-600",
            orange: "bg-orange-50 text-orange-600",
            purple: "bg-purple-50 text-purple-600",
          };

          return (
            <Card
              key={stat.title}
              className="hover:shadow-md transition-shadow cursor-pointer"
            >
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-slate-500">{stat.title}</p>
                    <p className="text-2xl font-bold text-slate-900 mt-1">
                      {stat.value}
                    </p>
                  </div>
                  <div
                    className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}
                  >
                    <Icon className="h-5 w-5" />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索接口名称、描述或端点..."
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex gap-2">
              <Select value={selectedCategory} onValueChange={handleCategoryChange}>
                <SelectTrigger className="h-9 w-40 cursor-pointer">
                  <SelectValue placeholder="全部分类" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="all">全部分类</SelectItem>
                    {categories.map((cat) => (
                      <SelectItem key={cat.id} value={cat.id}>
                        {cat.name} ({cat.apiCount})
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
              <Select value={selectedStatus} onValueChange={handleStatusChange}>
                <SelectTrigger className="h-9 w-32 cursor-pointer">
                  <SelectValue placeholder="全部状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="all">全部状态</SelectItem>
                    <SelectItem value="active">已启用</SelectItem>
                    <SelectItem value="inactive">已停用</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* APIs Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">接口列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
            </div>
          ) : apis.length === 0 ? (
            <div className="text-center py-12">
              <Settings className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">
                没有找到接口
              </h3>
              <p className="text-slate-500 mb-4">
                尝试调整搜索条件或添加新接口
              </p>
              <Button className="gap-2 cursor-pointer" onClick={handleAddApi}>
                <Plus className="h-4 w-4" />
                添加接口
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        接口信息
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        别名
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        方法
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        分类
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        定价
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        状态
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        调用次数
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        操作
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {apis.map((api) => (
                      <tr
                        key={api.id}
                        className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                      >
                        <td className="py-3 px-4">
                          <div>
                            <p className="text-sm font-medium text-slate-900">
                              {api.name}
                            </p>
                            <p className="text-xs text-slate-500 mt-1">
                              {api.description}
                            </p>
                            <p
                              className="text-xs text-slate-400 mt-1 font-mono truncate max-w-xs"
                              title={api.endpoint}
                            >
                              {api.endpoint}
                            </p>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <code className="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700">
                            {api.alias || "-"}
                          </code>
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              api.method === "GET"
                                ? "bg-green-50 text-green-700 border-green-200"
                                : api.method === "POST"
                                  ? "bg-blue-50 text-blue-700 border-blue-200"
                                  : api.method === "PUT"
                                    ? "bg-orange-50 text-orange-700 border-orange-200"
                                    : "bg-red-50 text-red-700 border-red-200"
                            }
                          >
                            {api.method}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {api.category.name}
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {api.pricing}
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              api.isActive
                                ? "bg-green-50 text-green-700 border-green-200 cursor-pointer"
                                : "bg-gray-50 text-gray-700 border-gray-200 cursor-pointer"
                            }
                            onClick={() => handleToggleStatus(api.id)}
                          >
                            {api.isActive ? "已启用" : "已停用"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {api.callCount.toLocaleString()}
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 cursor-pointer"
                              onClick={() => handleEditApi(api)}
                              title="编辑"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                              onClick={() => handleDelete(api.id)}
                              title="删除"
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {pagination && (
                <div className="mt-4">
                  <Pagination
                    currentPage={pagination.page}
                    totalPages={pagination.totalPages}
                    total={pagination.total}
                    pageSize={pagination.limit}
                    onPageChange={handlePageChange}
                    onPageSizeChange={handlePageSizeChange}
                  />
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Categories Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">接口分类</CardTitle>
            <Button
              variant="outline"
              size="sm"
              className="gap-2 cursor-pointer"
              onClick={handleAddCategory}
            >
              <Plus className="h-4 w-4" />
              添加分类
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3 lg:grid-cols-4">
            {categories.map((category) => (
              <div
                key={category.id}
                className="flex items-center justify-between p-4 rounded-lg border border-slate-200 hover:border-blue-300 hover:bg-blue-50/50 transition-colors cursor-pointer"
              >
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center">
                    <FolderOpen className="h-5 w-5 text-white" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-slate-900">
                      {category.name}
                    </p>
                    <p className="text-xs text-slate-500">
                      {category.apiCount} 个接口
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* API Form Dialog */}
      <Dialog open={showApiForm} onOpenChange={setShowApiForm}>
        <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{editingApi ? "编辑接口" : "添加接口"}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto pr-2 -mr-2">
            <ApiForm
              api={editingApi || undefined}
              categories={categories.map((c) => ({ id: c.id, name: c.name }))}
              onClose={() => setShowApiForm(false)}
              onSuccess={handleApiFormSuccess}
              formId="api-form"
            />
          </div>
          <DialogFooter className="border-t pt-4 mt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowApiForm(false)}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button type="submit" form="api-form" className="cursor-pointer">
              {editingApi ? "保存" : "添加"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Category Form Dialog */}
      <Dialog open={showCategoryForm} onOpenChange={setShowCategoryForm}>
        <DialogContent className="max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>
              {editingCategory ? "编辑分类" : "添加分类"}
            </DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto">
            <CategoryForm
              category={editingCategory || undefined}
              onClose={() => setShowCategoryForm(false)}
              onSuccess={handleCategoryFormSuccess}
              formId="category-form"
            />
          </div>
          <DialogFooter className="border-t pt-4 mt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowCategoryForm(false)}
              className="cursor-pointer"
            >
              取消
            </Button>
            <Button type="submit" form="category-form" className="cursor-pointer">
              {editingCategory ? "保存" : "添加"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
