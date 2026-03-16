"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { 
  FileText, 
  Search, 
  Filter,
  Download,
  RefreshCw,
  User,
  Settings,
  Shield,
  AlertCircle,
  Info,
  Clock
} from "lucide-react";

const auditLogs = [
  {
    id: 1,
    timestamp: "2024-03-13 14:23:45",
    user: "admin@example.com",
    action: "用户登录",
    resource: "用户系统",
    details: "管理员登录成功",
    ip: "192.168.1.100",
    status: "success",
    level: "info",
  },
  {
    id: 2,
    timestamp: "2024-03-13 14:20:12",
    user: "admin@example.com",
    action: "API 配置修改",
    resource: "GPT-4 对话 API",
    details: "修改了 API 定价从 0.01 到 0.02 积分/次",
    ip: "192.168.1.100",
    status: "success",
    level: "warning",
  },
  {
    id: 3,
    timestamp: "2024-03-13 13:45:30",
    user: "user@example.com",
    action: "API 调用失败",
    resource: "天气查询 API",
    details: "调用失败：接口维护中",
    ip: "192.168.1.105",
    status: "error",
    level: "error",
  },
  {
    id: 4,
    timestamp: "2024-03-13 12:30:15",
    user: "admin@example.com",
    action: "用户权限变更",
    resource: "用户管理",
    details: "将用户 user2@example.com 升级为管理员",
    ip: "192.168.1.100",
    status: "success",
    level: "warning",
  },
  {
    id: 5,
    timestamp: "2024-03-13 11:20:45",
    user: "system",
    action: "系统备份",
    resource: "数据库",
    details: "自动备份完成",
    ip: "-",
    status: "success",
    level: "info",
  },
  {
    id: 6,
    timestamp: "2024-03-13 10:15:20",
    user: "user3@example.com",
    action: "账户充值",
    resource: "支付系统",
    details: "充值 ¥200.00",
    ip: "192.168.1.110",
    status: "success",
    level: "info",
  },
  {
    id: 7,
    timestamp: "2024-03-13 09:45:10",
    user: "admin@example.com",
    action: "API 下线",
    resource: "OCR 文字识别 API",
    details: "API 进入维护模式",
    ip: "192.168.1.100",
    status: "success",
    level: "warning",
  },
  {
    id: 8,
    timestamp: "2024-03-12 18:30:25",
    user: "user4@example.com",
    action: "登录失败",
    resource: "用户系统",
    details: "密码错误次数过多",
    ip: "192.168.1.115",
    status: "error",
    level: "error",
  },
];

const logLevels = [
  { name: "全部", count: 1234 },
  { name: "信息", count: 987, color: "blue" },
  { name: "警告", count: 187, color: "orange" },
  { name: "错误", count: 60, color: "red" },
];

export default function AuditLogsPage() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">审计日志</h1>
          <p className="text-slate-500 mt-1">查看系统操作记录和安全日志</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>
          <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
            <Download className="h-4 w-4" />
            导出
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        {logLevels.map((level) => (
          <Card key={level.name} className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-slate-500">{level.name}日志</p>
                  <p className="text-2xl font-bold text-slate-900 mt-1">{level.count}</p>
                </div>
                <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${
                  level.color === "blue" ? "bg-blue-50" :
                  level.color === "orange" ? "bg-orange-50" :
                  level.color === "red" ? "bg-red-50" :
                  "bg-slate-50"
                }`}>
                  {level.color === "blue" ? (
                    <Info className="h-5 w-5 text-blue-600" />
                  ) : level.color === "orange" ? (
                    <AlertCircle className="h-5 w-5 text-orange-600" />
                  ) : level.color === "red" ? (
                    <AlertCircle className="h-5 w-5 text-red-600" />
                  ) : (
                    <FileText className="h-5 w-5 text-slate-600" />
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索用户、操作或资源..."
                className="pl-10"
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
                <Clock className="h-4 w-4" />
                时间范围
              </Button>
              <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
                <Filter className="h-4 w-4" />
                筛选
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Logs Table */}
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
                </tr>
              </thead>
              <tbody>
                {auditLogs.map((log) => (
                  <tr key={log.id} className="border-b border-slate-100 hover:bg-slate-50 transition-colors">
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-slate-400" />
                        <span className="text-sm text-slate-600">{log.timestamp}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <User className="h-4 w-4 text-slate-400" />
                        <span className="text-sm text-slate-900">{log.user}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-sm text-slate-900">{log.action}</td>
                    <td className="py-3 px-4 text-sm text-slate-600">{log.resource}</td>
                    <td className="py-3 px-4">
                      <span className="text-sm text-slate-600 max-w-xs truncate block">
                        {log.details}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <code className="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700">
                        {log.ip}
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
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
