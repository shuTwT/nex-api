import { useEffect, useState } from "react";
import { Card, Col, Row, Tag, Typography } from "antd";
import { Zap, BarChart3, DollarSign, Activity } from "lucide-react";
import { api, responseData } from "@/lib/api";
import { UsageTrendChart } from "@/components/usage-trend-chart";
import { ConsolePageLoading } from "@/components/console-page-loading";

interface DashboardStats { monthlyCreditsUsed: number; apiCalls: number; accountBalance: number; activeApis: number; }
interface ActivityRecord { id: string; apiName: string; apiAlias: string; credits: number; status: string; createdAt: Date; }
interface TopApi { name: string; calls: number; percentage: number; }

export default function ConsoleDashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [activities, setActivities] = useState<ActivityRecord[]>([]);
  const [topApis, setTopApis] = useState<TopApi[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  async function loadData() {
    setIsLoading(true);
    const [statsResult, activityResult, topApisResult] = await Promise.all([api.dashboard_stats_route_get(), api.dashboard_activity_route_get(), api.dashboard_top_apis_route_get()]);
    const statsData = responseData<DashboardStats>(statsResult);
    const activityData = responseData<ActivityRecord[]>(activityResult);
    const topApisData = responseData<TopApi[]>(topApisResult);
    if (statsData) setStats(statsData);
    if (activityData) setActivities(activityData);
    if (topApisData) setTopApis(topApisData);
    setIsLoading(false);
  }
  useEffect(() => { loadData(); }, []);

  const statsCards = [
    { title: "本月积分使用", value: stats?.monthlyCreditsUsed.toLocaleString() || 0, icon: Zap },
    { title: "API 调用次数", value: stats?.apiCalls.toLocaleString() || 0, icon: BarChart3 },
    { title: "账户余额", value: `${stats?.accountBalance.toLocaleString() || 0} 积分`, icon: DollarSign },
    { title: "活跃接口数", value: stats?.activeApis || 0, icon: Activity },
  ];

  if (isLoading) return <ConsolePageLoading />;

  return <div className="flex flex-col gap-6">
    <div><Typography.Title level={2}>概览</Typography.Title><Typography.Text type="secondary">欢迎回来，这是您的账户概览</Typography.Text></div>
    <Row gutter={[24, 24]}>
      {statsCards.map((stat) => { const Icon = stat.icon; return <Col key={stat.title} xs={24} md={12} xl={6}><Card hoverable><Icon size={24} /><Typography.Text type="secondary" className="block mt-4">{stat.title}</Typography.Text><Typography.Title level={3}>{stat.value}</Typography.Title></Card></Col>; })}
    </Row>
    <Row gutter={[24, 24]}><Col xs={24} lg={16}><Card title="用量趋势"><UsageTrendChart className="h-64" /></Card></Col><Col xs={24} lg={8}><Card title="热门接口">{topApis.length === 0 ? <Typography.Text type="secondary">暂无数据</Typography.Text> : topApis.map((item, index) => <div key={item.name} className="flex items-center justify-between py-2"><span>{index + 1}. {item.name}</span><Typography.Text type="secondary">{item.calls.toLocaleString()} 次调用 · {item.percentage}%</Typography.Text></div>)}</Card></Col></Row>
    <Card title="最近活动">{activities.length === 0 ? <Typography.Text type="secondary">暂无活动记录</Typography.Text> : activities.map((activity) => <div key={activity.id} className="flex items-center justify-between py-3 border-b"><div><Typography.Text strong>{activity.apiName}</Typography.Text><br /><Typography.Text type="secondary">{new Date(activity.createdAt).toLocaleString("zh-CN")}</Typography.Text></div><div className="flex items-center gap-3"><Tag color={activity.status === "success" ? "success" : "error"}>{activity.status === "success" ? "成功" : "失败"}</Tag><Typography.Text>{activity.credits} 积分</Typography.Text></div></div>)}</Card>
  </div>;
}
