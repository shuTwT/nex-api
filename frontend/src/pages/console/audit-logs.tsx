import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  FileText,
  Search,
  Download,
  RefreshCw,
  User,
  AlertCircle,
  Info,
  Clock,
  Trash2,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { DeleteAuditLogDialog } from "@/components/delete-audit-log-dialog";
import { Pagination } from "@/components/pagination";

interface AuditLog {
  id: string;
  userId: string | null;
  user: {
    id: string;
    name: string | null;
    email: string | null;
  } | null;
  action: string;
  resource: string;
  details: string | null;
  ipAddress: string | null;
  userAgent: string | null;
  level: string;
  status: string;
  metadata: string | null;
  createdAt: string;
}

interface AuditLogStats {
  totalLogs: number;
  infoLogs: number;
  warningLogs: number;
  errorLogs: number;
  successLogs: number;
  failedLogs: number;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

function toRFC3339(value: string): string {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [stats, setStats] = useState<AuditLogStats | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [levelFilter, setLevelFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [startDateInput, setStartDateInput] = useState("");
  const [endDateInput, setEndDateInput] = useState("");
  const [appliedSearch, setAppliedSearch] = useState("");
  const [appliedLevel, setAppliedLevel] = useState<string>("all");
  const [appliedStatus, setAppliedStatus] = useState<string>("all");
  const [appliedStartDate, setAppliedStartDate] = useState("");
  const [appliedEndDate, setAppliedEndDate] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(true);
  const [deletingAuditLog, setDeletingAuditLog] = useState<AuditLog | null>(null);

  const loadLogs = useCallback(async () => {
    setIsLoading(true);
    const query: Record<string, string | number | boolean> = {
      search: appliedSearch,
      level: appliedLevel,
      status: appliedStatus,
      startDate: toRFC3339(appliedStartDate),
      endDate: toRFC3339(appliedEndDate),
      page: currentPage,
      limit: pageSize,
    };
    const result = await api.audit_logs_route_get(query);

    if (result.success) {
      const data = responseData<AuditLog[]>(result);
      if (data) setLogs(data);
      if (result.pagination) {
        setPagination(result.pagination);
      }
    }
    setIsLoading(false);
  }, [currentPage, pageSize, appliedSearch, appliedLevel, appliedStatus, appliedStartDate, appliedEndDate]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  async function loadStats() {
    const result = await api.audit_logs_stats_route_get();
    const data = responseData<AuditLogStats>(result);
    if (data) setStats(data);
  }

  useEffect(() => {
    loadStats();
  }, []);

  function handlePageChange(page: number) {
    setCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    setPageSize(size);
    setCurrentPage(1);
  }

  function handleQuery() {
    setAppliedSearch(searchInput);
    setAppliedLevel(levelFilter);
    setAppliedStatus(statusFilter);
    setAppliedStartDate(startDateInput);
    setAppliedEndDate(endDateInput);
    setCurrentPage(1);
  }

  function handleReset() {
    setSearchInput("");
    setLevelFilter("all");
    setStatusFilter("all");
    setStartDateInput("");
    setEndDateInput("");
    setAppliedSearch("");
    setAppliedLevel("all");
    setAppliedStatus("all");
    setAppliedStartDate("");
    setAppliedEndDate("");
    setCurrentPage(1);
  }

  function handleDeleteAuditLog(log: AuditLog) {
    setDeletingAuditLog(log);
  }

  function handleFormSuccess() {
    loadLogs();
    loadStats();
  }

  async function handleExport() {
    const query: Record<string, string | number | boolean> = {
      level: levelFilter,
      status: statusFilter,
      startDate: toRFC3339(startDateInput),
      endDate: toRFC3339(endDateInput),
    };
    const result = await api.audit_logs_export_route_get(query);

    if (result.success) {
      const data = responseData<string>(result);
      if (data !== null) {
        const blob = new Blob([data], { type: "text/csv;charset=utf-8;" });
        const link = document.createElement("a");
        const url = URL.createObjectURL(blob);
        link.setAttribute("href", url);
        link.setAttribute("download", `audit-logs-${new Date().toISOString()}.csv`);
        link.style.visibility = "hidden";
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      }
    } else {
      alert(result.error || "导出失败");
    }
  }

  const statsCards = [
    {
      title: "总日志",
      value: stats?.totalLogs || 0,
      icon: FileText,
      color: "blue",
    },
    {
      title: "信息",
      value: stats?.infoLogs || 0,
      icon: Info,
      color: "blue",
    },
    {
      title: "警告",
      value: stats?.warningLogs || 0,
      icon: AlertCircle,
      color: "orange",
    },
    {
      title: "错误",
      value: stats?.errorLogs || 0,
      icon: AlertCircle,
      color: "red",
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">审计日志</h1>
          <p className="text-slate-500 mt-1">查看系统操作记录和安全日志</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="gap-2 cursor-pointer" onClick={loadLogs}>
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          <Button variant="outline" size="sm" className="gap-2 cursor-pointer" onClick={handleExport}>
            <Download className="h-4 w-4" />
            导出
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            orange: "bg-orange-50 text-orange-600",
            red: "bg-red-50 text-red-600",
          };

          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-slate-500">{stat.title}日志</p>
                    <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                  </div>
                  <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
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
                placeholder="搜索操作、资源或详情..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={levelFilter} onValueChange={setLevelFilter}>
              <SelectTrigger className="w-[110px]">
                <SelectValue placeholder="级别" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部级别</SelectItem>
                  <SelectItem value="info">信息</SelectItem>
                  <SelectItem value="warning">警告</SelectItem>
                  <SelectItem value="error">错误</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[110px]">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="success">成功</SelectItem>
                  <SelectItem value="error">失败</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <div className="flex-1 min-w-[180px]">
              <Input
                type="datetime-local"
                value={startDateInput}
                onChange={(e) => setStartDateInput(e.target.value)}
              />
            </div>
            <div className="flex-1 min-w-[180px]">
              <Input
                type="datetime-local"
                value={endDateInput}
                onChange={(e) => setEndDateInput(e.target.value)}
              />
            </div>
            <Button
              size="sm"
              onClick={handleQuery}
              className="cursor-pointer"
            >
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

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
        </div>
      ) : logs.length === 0 ? (
        <Card>
          <CardContent className="p-12">
            <div className="text-center">
              <FileText className="h-12 w-12 text-slate-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">没有找到审计日志</h3>
              <p className="text-slate-500 mb-4">暂无审计日志记录</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">日志列表</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-slate-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">时间</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">用户</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">操作</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">资源</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">详情</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">IP 地址</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">级别</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">状态</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-slate-600">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {logs.map((log) => (
                      <tr key={log.id} className="border-b border-slate-100 hover:bg-slate-50 transition-colors">
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <Clock className="h-4 w-4 text-slate-400" />
                            <span className="text-sm text-slate-600">{new Date(log.createdAt).toLocaleString("zh-CN")}</span>
                          </div>
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex items-center gap-2">
                            <User className="h-4 w-4 text-slate-400" />
                            <span className="text-sm text-slate-900">{log.user?.email || log.user?.name || "系统"}</span>
                          </div>
                        </td>
                        <td className="py-3 px-4 text-sm text-slate-900">{log.action}</td>
                        <td className="py-3 px-4 text-sm text-slate-600">{log.resource}</td>
                        <td className="py-3 px-4">
                          <span className="text-sm text-slate-600 max-w-xs truncate block">
                            {log.details || "-"}
                          </span>
                        </td>
                        <td className="py-3 px-4">
                          <code className="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700">
                            {log.ipAddress || "-"}
                          </code>
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              log.level === "info"
                                ? "bg-blue-50 text-blue-700 border-blue-200"
                                : log.level === "warning"
                                ? "bg-orange-50 text-orange-700 border-orange-200"
                                : "bg-red-50 text-red-700 border-red-200"
                            }
                          >
                            {log.level === "info" ? "信息" : log.level === "warning" ? "警告" : "错误"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4">
                          <Badge
                            variant="outline"
                            className={
                              log.status === "success"
                                ? "bg-green-50 text-green-700 border-green-200"
                                : "bg-red-50 text-red-700 border-red-200"
                            }
                          >
                            {log.status === "success" ? "成功" : "失败"}
                          </Badge>
                        </td>
                        <td className="py-3 px-4">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDeleteAuditLog(log)}
                            className="text-red-600 hover:text-red-700 hover:bg-red-50 cursor-pointer"
                            title="删除"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

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

      {deletingAuditLog && (
        <DeleteAuditLogDialog
          open={!!deletingAuditLog}
          onOpenChange={(open) => !open && setDeletingAuditLog(null)}
          auditLogId={deletingAuditLog.id}
          auditLogAction={deletingAuditLog.action}
          onSuccess={handleFormSuccess}
        />
      )}
    </div>
  );
}
