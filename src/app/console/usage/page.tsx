"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { 
  TrendingUp,
  Zap,
  Calendar,
  Clock,
  Gift,
  Activity,
  BarChart3
} from "lucide-react";
import { api } from "@/lib/api-client";

interface UsageStats {
  freeCredits: number;
  totalUsage: number;
  todayUsage: number;
  last7DaysUsage: number;
  last30DaysUsage: number;
  todayHourlyUsage: number[];
  last7DaysDailyUsage: number[];
  last30DaysDailyUsage: number[];
}

export default function UsagePage() {
  const [stats, setStats] = useState<UsageStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [today] = useState(() => Date.now());

  useEffect(() => {
    async function loadStats() {
      setIsLoading(true);
      const result = await api.get("/api/usage");
      if (result.success && result.data) {
        setStats(result.data);
      }
      setIsLoading(false);
    }
    loadStats();
  }, []);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-slate-500">无法加载用量统计</p>
      </div>
    );
  }

  const todayHours = Array.from({ length: 24 }, (_, i) => i);
  const last7Days = Array.from({ length: 7 }, (_, i) => i);
  const last30Days = Array.from({ length: 30 }, (_, i) => i);

  const todayHourlyData = todayHours.map((hour, index) => ({
    label: `${index}:00`,
    value: stats.todayHourlyUsage[index] || 0,
  }));

  const last7DaysData = last7Days.map((day, index) => ({
    label: new Date(today - (6 - index) * 24 * 60 * 60 * 1000).toLocaleDateString("zh-CN", { weekday: "short" }),
    value: stats.last7DaysDailyUsage[index] || 0,
  }));

  const last30DaysData = last30Days.map((day, index) => ({
    label: new Date(today - (29 - index) * 24 * 60 * 60 * 1000).toLocaleDateString("zh-CN", { month: "short", day: "numeric" }),
    value: stats.last30DaysDailyUsage[index] || 0,
  }));

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">用量统计</h1>
        <p className="text-slate-500 mt-1">查看您的 API 使用情况</p>
      </div>

      {/* Credits Overview */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="border-2 border-blue-200 bg-gradient-to-br from-blue-50 to-white">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2 mb-2">
                <Gift className="h-5 w-5 text-blue-600" />
                <span className="text-sm font-medium text-blue-700">免费额度</span>
              </div>
                <p className="text-3xl font-bold text-blue-900">{stats.freeCredits.toLocaleString()}</p>
                <p className="text-xs text-blue-600 mt-1">剩余可用积分</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-2 border-purple-200 bg-gradient-to-br from-purple-50 to-white">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2 mb-2">
                <Zap className="h-5 w-5 text-purple-600" />
                <span className="text-sm font-medium text-purple-700">付费额度</span>
              </div>
                <p className="text-3xl font-bold text-purple-900">0</p>
                <p className="text-xs text-purple-600 mt-1">已购买的额外积分</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Usage Summary Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card className="hover:shadow-md transition-shadow cursor-pointer">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">今日消耗</p>
                <p className="text-2xl font-bold text-slate-900 mt-1">{stats.todayUsage.toLocaleString()}</p>
              </div>
              <div className="h-10 w-10 rounded-lg bg-orange-50 flex items-center justify-center">
                <Activity className="h-5 w-5 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow cursor-pointer">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">7 日消耗</p>
                <p className="text-2xl font-bold text-slate-900 mt-1">{stats.last7DaysUsage.toLocaleString()}</p>
              </div>
              <div className="h-10 w-10 rounded-lg bg-blue-50 flex items-center justify-center">
                <Calendar className="h-5 w-5 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="hover:shadow-md transition-shadow cursor-pointer">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">30 日消耗</p>
                <p className="text-2xl font-bold text-slate-900 mt-1">{stats.last30DaysUsage.toLocaleString()}</p>
              </div>
              <div className="h-10 w-10 rounded-lg bg-purple-50 flex items-center justify-center">
                <TrendingUp className="h-5 w-5 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Today's Hourly Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Clock className="h-5 w-5 text-blue-600" />
            今日消耗（按小时）
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-64 flex items-end gap-1">
              {todayHourlyData.map((item, index) => {
                const maxValue = Math.max(...todayHourlyData.map(d => d.value));
                const height = maxValue > 0 ? (item.value / maxValue) * 100 : 0;
                
                return (
                  <div
                    key={index}
                    className="flex-1 bg-gradient-to-t from-blue-600 to-blue-400 rounded-t cursor-pointer hover:from-blue-700 hover:to-blue-500 transition-all relative group"
                    style={{ height: `${height}%`, minHeight: item.value > 0 ? "4px" : "0" }}
                  >
                    <div className="absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <div className="bg-slate-900 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                        {item.value} 次
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex justify-between text-xs text-slate-500">
              {todayHourlyData.filter((_, i) => i % 6 === 0).map((item, index) => (
                <span key={index}>{item.label}</span>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 7 Days Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Calendar className="h-5 w-5 text-blue-600" />
            7 日消耗（按天）
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-64 flex items-end gap-2">
              {last7DaysData.map((item, index) => {
                const maxValue = Math.max(...last7DaysData.map(d => d.value));
                const height = maxValue > 0 ? (item.value / maxValue) * 100 : 0;
                
                return (
                  <div
                    key={index}
                    className="flex-1 bg-gradient-to-t from-cyan-600 to-cyan-400 rounded-t cursor-pointer hover:from-cyan-700 hover:to-cyan-500 transition-all relative group"
                    style={{ height: `${height}%`, minHeight: item.value > 0 ? "4px" : "0" }}
                  >
                    <div className="absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <div className="bg-slate-900 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                        {item.value} 次
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex justify-between text-xs text-slate-500">
              {last7DaysData.map((item, index) => (
                <span key={index}>{item.label}</span>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 30 Days Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-blue-600" />
            30 日消耗（按天）
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-64 flex items-end gap-1">
              {last30DaysData.map((item, index) => {
                const maxValue = Math.max(...last30DaysData.map(d => d.value));
                const height = maxValue > 0 ? (item.value / maxValue) * 100 : 0;
                
                return (
                  <div
                    key={index}
                    className="flex-1 bg-gradient-to-t from-purple-600 to-purple-400 rounded-t cursor-pointer hover:from-purple-700 hover:to-purple-500 transition-all relative group"
                    style={{ height: `${height}%`, minHeight: item.value > 0 ? "4px" : "0" }}
                  >
                    <div className="absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <div className="bg-slate-900 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                        {item.value} 次
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex justify-between text-xs text-slate-500">
              {last30DaysData.filter((_, i) => i % 5 === 0).map((item, index) => (
                <span key={index}>{item.label}</span>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Total Usage */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-blue-600" />
            总计使用量
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <p className="text-6xl font-bold text-slate-900">{stats.totalUsage.toLocaleString()}</p>
              <p className="text-sm text-slate-500 mt-2">总 API 调用次数</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
