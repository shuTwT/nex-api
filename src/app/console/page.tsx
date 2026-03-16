"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { 
  TrendingUp, 
  TrendingDown, 
  Zap, 
  Users, 
  BarChart3, 
  DollarSign,
  ArrowUpRight,
  ArrowDownRight
} from "lucide-react";

const stats = [
  {
    title: "本月积分使用",
    value: "2,345",
    change: "+12.5%",
    trend: "up",
    icon: Zap,
    color: "blue",
  },
  {
    title: "API 调用次数",
    value: "15,234",
    change: "+8.2%",
    trend: "up",
    icon: BarChart3,
    color: "cyan",
  },
  {
    title: "账户余额",
    value: "¥ 128.50",
    change: "-3.1%",
    trend: "down",
    icon: DollarSign,
    color: "green",
  },
  {
    title: "活跃接口数",
    value: "23",
    change: "+2",
    trend: "up",
    icon: Users,
    color: "purple",
  },
];

const recentActivity = [
  {
    id: 1,
    api: "GPT-4 对话 API",
    action: "调用成功",
    time: "2 分钟前",
    status: "success",
    credits: "0.02",
  },
  {
    id: 2,
    api: "IP 地址查询",
    action: "调用成功",
    time: "5 分钟前",
    status: "success",
    credits: "0.01",
  },
  {
    id: 3,
    api: "天气查询",
    action: "调用失败",
    time: "10 分钟前",
    status: "error",
    credits: "0",
  },
  {
    id: 4,
    api: "OCR 文字识别",
    action: "调用成功",
    time: "15 分钟前",
    status: "success",
    credits: "0.05",
  },
  {
    id: 5,
    api: "汇率查询",
    action: "调用成功",
    time: "20 分钟前",
    status: "success",
    credits: "0.01",
  },
];

const topAPIs = [
  { name: "GPT-4 对话 API", calls: 1234, percentage: 35 },
  { name: "IP 地址查询", calls: 987, percentage: 28 },
  { name: "天气查询", calls: 654, percentage: 19 },
  { name: "OCR 文字识别", calls: 432, percentage: 12 },
  { name: "汇率查询", calls: 234, percentage: 6 },
];

export default function ConsoleDashboard() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">概览</h1>
        <p className="text-slate-500 mt-1">欢迎回来，这是您的账户概览</p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            cyan: "bg-cyan-50 text-cyan-600",
            green: "bg-green-50 text-green-600",
            purple: "bg-purple-50 text-purple-600",
          };
          
          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className={`h-12 w-12 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                    <Icon className="h-6 w-6" />
                  </div>
                  <div className={`flex items-center gap-1 text-sm ${
                    stat.trend === "up" ? "text-green-600" : "text-red-600"
                  }`}>
                    {stat.trend === "up" ? (
                      <ArrowUpRight className="h-4 w-4" />
                    ) : (
                      <ArrowDownRight className="h-4 w-4" />
                    )}
                    <span>{stat.change}</span>
                  </div>
                </div>
                <div className="mt-4">
                  <p className="text-sm text-slate-500">{stat.title}</p>
                  <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Charts and Activity */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Usage Chart */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">用量趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-slate-50 rounded-lg">
              <div className="text-center">
                <BarChart3 className="h-12 w-12 text-slate-300 mx-auto mb-3" />
                <p className="text-sm text-slate-500">图表区域</p>
                <p className="text-xs text-slate-400 mt-1">集成 Chart.js 或 Recharts</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Top APIs */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">热门接口</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topAPIs.map((api, index) => (
                <div key={api.name} className="flex items-center gap-3">
                  <div className="h-8 w-8 rounded-full bg-slate-100 flex items-center justify-center text-sm font-medium text-slate-600">
                    {index + 1}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-slate-900 truncate">{api.name}</p>
                    <p className="text-xs text-slate-500">{api.calls.toLocaleString()} 次调用</p>
                  </div>
                  <div className="text-sm text-slate-600">{api.percentage}%</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">最近活动</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentActivity.map((activity) => (
              <div key={activity.id} className="flex items-center gap-4 py-3 border-b border-slate-100 last:border-0">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-slate-900">{activity.api}</p>
                  <p className="text-xs text-slate-500 mt-1">{activity.action}</p>
                </div>
                <div className="flex items-center gap-3">
                  <Badge 
                    variant={activity.status === "success" ? "default" : "destructive"}
                    className={activity.status === "success" ? "bg-green-50 text-green-700 border-green-200" : ""}
                  >
                    {activity.status === "success" ? "成功" : "失败"}
                  </Badge>
                  <div className="text-right">
                    <p className="text-sm font-medium text-slate-900">{activity.credits} 积分</p>
                    <p className="text-xs text-slate-500">{activity.time}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
