"use client";

import { useEffect, useState, useCallback, useTransition } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
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
  Plug,
  Activity,
  Pause,
  Database,
} from "lucide-react";
import { api } from "@/lib/api-client";
import { Pagination } from "@/components/pagination";
import { McpFormDialog } from "./components/mcp-form-dialog";
import type { McpServiceData } from "./components/mcp-form-dialog";

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

interface McpStats {
  totalServices: number;
  activeServices: number;
  inactiveServices: number;
  totalCalls: number;
}

const TYPE_LABELS: Record<string, string> = {
  stdio: "stdio",
  sse: "SSE",
  streamableHttp: "Streamable HTTP",
};

export default function McpServicesPage() {
  const [services, setServices] = useState<McpServiceData[]>([]);
  const [stats, setStats] = useState<McpStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [typeFilter, setTypeFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedType, setAppliedType] = useState("all");
  const [appliedStatus, setAppliedStatus] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [isPending, startTransition] = useTransition();
  const [showForm, setShowForm] = useState(false);
  const [editingService, setEditingService] = useState<McpServiceData | null>(null);

  const loadServices = useCallback(async () => {
    setIsLoading(true);
    const result = await api.paginated("/api/mcp-services", {
      type: appliedType,
      search: appliedSearch,
      status: appliedStatus,
      page: currentPage,
      limit: pageSize,
    });

    if (result.success && result.data) {
      setServices(result.data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedType, appliedStatus]);

  useEffect(() => {
    loadServices();
  }, [loadServices]);

  useEffect(() => {
    loadStats();
  }, []);

  async function loadStats() {
    const result = await api.get("/api/mcp-services/stats");
    if (result.success && result.data) {
      setStats(result.data);
    }
  }

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedType(typeFilter);
    setAppliedStatus(statusFilter);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setTypeFilter("all");
    setStatusFilter("all");
    setAppliedSearch("");
    setAppliedType("all");
    setAppliedStatus("all");
    setCurrentPage(1);
  }

  async function handleToggleStatus(id: string) {
    startTransition(async () => {
      const result = await api.put(`/api/mcp-services/${id}/toggle`);
      if (result.success) {
        loadServices();
        loadStats();
      } else {
        alert(result.error || "切换状态失败");
      }
    });
  }

  async function handleDelete(id: string) {
    if (!confirm("确定要删除这个 MCP 服务吗？此操作无法撤销。")) {
      return;
    }

    startTransition(async () => {
      const result = await api.delete(`/api/mcp-services/${id}`);
      if (result.success) {
        loadServices();
        loadStats();
      } else {
        alert(result.error || "删除失败");
      }
    });
  }

  function handleAdd() {
    setEditingService(null);
    setShowForm(true);
  }

  function handleEdit(service: McpServiceData) {
    setEditingService(service);
    setShowForm(true);
  }

  async function handleFormSuccess() {
    await loadServices();
    await loadStats();
  }

  const statsCards = [
    {
      title: "服务总数",
      value: stats?.totalServices || 0,
      icon: Plug,
      color: "blue",
    },
    {
      title: "活跃服务",
      value: stats?.activeServices || 0,
      icon: Activity,
      color: "green",
    },
    {
      title: "停用服务",
      value: stats?.inactiveServices || 0,
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">MCP 服务管理</h1>
          <p className="text-slate-500 mt-1">管理 MCP（Model Context Protocol）服务</p>
        </div>
        <Button className="gap-2 cursor-pointer" onClick={handleAdd}>
          <Plus className="h-4 w-4" />
          添加服务
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses: Record<string, string> = {
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
                    className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color]}`}
                  >
                    <Icon className="h-5 w-5" />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索服务名称或标识..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={typeFilter} onValueChange={setTypeFilter}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="全部类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部类型</SelectItem>
                  <SelectItem value="stdio">stdio</SelectItem>
                  <SelectItem value="sse">SSE</SelectItem>
                  <SelectItem value="streamableHttp">Streamable HTTP</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[130px]">
                <SelectValue placeholder="全部状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="active">已启用</SelectItem>
                  <SelectItem value="inactive">已停用</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button size="sm" onClick={handleQuery} className="cursor-pointer">
              查询
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleReset}
              className="cursor-pointer"
            >
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">服务列表</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
            </div>
          ) : services.length === 0 ? (
            <div className="text-center py-12">
              <Plug className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">
                没有找到 MCP 服务
              </h3>
              <p className="text-slate-500 mb-4">
                尝试调整搜索条件或添加新服务
              </p>
              <Button className="gap-2 cursor-pointer" onClick={handleAdd}>
                <Plus className="h-4 w-4" />
                添加服务
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        服务信息
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        标识
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        类型
                      </th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">
                        端点
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
                    {services.map((svc) => (
                      <tr
                        key={svc.id}
                        className="border-b border-slate-100 hover:bg-slate-50 transition-colors"
                      >
                        <td className="py-3 px-4">
                          <div>
                            <p className="text-sm font-medium text-slate-900">
                              {svc.name}
                            </p>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <code className="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700">
                            {svc.identifier}
                          </code>
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              svc.type === "stdio"
                                ? "bg-gray-50 text-gray-700 border-gray-200"
                                : svc.type === "sse"
                                  ? "bg-blue-50 text-blue-700 border-blue-200"
                                  : "bg-purple-50 text-purple-700 border-purple-200"
                            }
                          >
                            {TYPE_LABELS[svc.type] || svc.type}
                          </Badge>
                        </td>
                        <td className="py-3 px-4">
                          <span
                            className="text-xs text-slate-500 font-mono truncate max-w-[200px] block"
                            title={svc.command || svc.endpoint || "-"}
                          >
                            {svc.type === "stdio"
                              ? svc.command || "-"
                              : svc.endpoint || "-"}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {svc.pricing}
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              svc.isActive
                                ? "bg-green-50 text-green-700 border-green-200 cursor-pointer"
                                : "bg-gray-50 text-gray-700 border-gray-200 cursor-pointer"
                            }
                            onClick={() => handleToggleStatus(svc.id)}
                          >
                            {svc.isActive ? "已启用" : "已停用"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-600">
                          {svc.callCount.toLocaleString()}
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 cursor-pointer"
                              onClick={() => handleEdit(svc)}
                              title="编辑"
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0 cursor-pointer text-red-600 hover:text-red-700 hover:bg-red-50"
                              onClick={() => handleDelete(svc.id)}
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

              <div className="mt-4">
                <Pagination
                  currentPage={pagination?.page ?? 1}
                  totalPages={pagination?.totalPages ?? 1}
                  total={pagination?.total ?? 0}
                  pageSize={pagination?.limit ?? pageSize}
                  onPageChange={handlePageChange}
                  onPageSizeChange={handlePageSizeChange}
                />
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <McpFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        service={editingService}
        onSuccess={handleFormSuccess}
      />
    </div>
  );
}
