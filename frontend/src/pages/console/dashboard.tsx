import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Zap,
  BarChart3,
  DollarSign,
  Activity,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { UsageTrendChart } from "@/components/usage-trend-chart";

interface DashboardStats {
  monthlyCreditsUsed: number;
  apiCalls: number;
  accountBalance: number;
  activeApis: number;
}

interface Activity {
  id: string;
  apiName: string;
  apiAlias: string;
  credits: number;
  status: string;
  createdAt: Date;
}

interface TopApi {
  name: string;
  calls: number;
  percentage: number;
}

export default function ConsoleDashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [topApis, setTopApis] = useState<TopApi[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  async function loadData() {
    setIsLoading(true);
    const [statsResult, activityResult, topApisResult] = await Promise.all([
      api.dashboard_stats_route_get(),
      api.dashboard_activity_route_get(),
      api.dashboard_top_apis_route_get(),
    ]);

    const statsData = responseData<DashboardStats>(statsResult);
    if (statsData) setStats(statsData);

    const activityData = responseData<Activity[]>(activityResult);
    if (activityData) setActivities(activityData);

    const topApisData = responseData<TopApi[]>(topApisResult);
    if (topApisData) setTopApis(topApisData);

    setIsLoading(false);
  }

  useEffect(() => {
    loadData();
  }, []);

  const statsCards = [
    {
      title: "本月积分使用",
      value: stats?.monthlyCreditsUsed.toLocaleString() || 0,
      icon: Zap,
      color: "blue",
    },
    {
      title: "API 调用次数",
      value: stats?.apiCalls.toLocaleString() || 0,
      icon: BarChart3,
      color: "cyan",
    },
    {
      title: "账户余额",
      value: `${stats?.accountBalance.toLocaleString() || 0} 积分`,
      icon: DollarSign,
      color: "green",
    },
    {
      title: "活跃接口数",
      value: stats?.activeApis || 0,
      icon: Activity,
      color: "purple",
    },
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">概览</h1>
        <p className="text-slate-500 mt-1">欢迎回来，这是您的账户概览</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {statsCards.map((stat) => {
          const Icon = stat.icon;
          const colorClasses = {
            blue: "bg-blue-50 text-blue-600",
            cyan: "bg-cyan-50 text-cyan-600",
            green: "bg-green-50 text-green-600",
            purple: "bg-purple-50 text-purple-600",
          };

          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className={`h-12 w-12 rounded-lg flex items-center justify-center ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                    <Icon className="h-6 w-6" />
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

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-lg">用量趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <UsageTrendChart className="h-64" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">热门接口</CardTitle>
          </CardHeader>
          <CardContent>
            {topApis.length === 0 ? (
              <div className="text-center py-8">
                <Activity className="h-8 w-8 text-slate-300 mx-auto mb-2" />
                <p className="text-sm text-slate-500">暂无数据</p>
              </div>
            ) : (
              <div className="space-y-4">
                {topApis.map((item, index) => (
                  <div key={item.name} className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-full bg-slate-100 flex items-center justify-center text-sm font-medium text-slate-600">
                      {index + 1}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-900 truncate">{item.name}</p>
                      <p className="text-xs text-slate-500">{item.calls.toLocaleString()} 次调用</p>
                    </div>
                    <div className="text-sm text-slate-600">{item.percentage}%</div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">最近活动</CardTitle>
        </CardHeader>
        <CardContent>
          {activities.length === 0 ? (
            <div className="text-center py-8">
              <Activity className="h-8 w-8 text-slate-300 mx-auto mb-2" />
              <p className="text-sm text-slate-500">暂无活动记录</p>
            </div>
          ) : (
            <div className="space-y-4">
              {activities.map((activity) => (
                <div key={activity.id} className="flex items-center gap-4 py-3 border-b border-slate-100 last:border-0">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-slate-900">{activity.apiName}</p>
                    <p className="text-xs text-slate-500 mt-1">
                      {new Date(activity.createdAt).toLocaleString("zh-CN")}
                    </p>
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
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
