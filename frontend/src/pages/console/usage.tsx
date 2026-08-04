import { useEffect, useState } from "react";
import { Card, Col, Progress, Row, Statistic, Typography } from "antd";
import {
  TrendingUp,
  Zap,
  Calendar,
  Clock,
  Gift,
  Activity,
  BarChart3,
} from "lucide-react";
import { api, responseData } from "@/lib/api";
import { ConsolePageLoading } from "@/components/console-page-loading";

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
interface UsageChartProps {
  title: string;
  icon: React.ReactNode;
  values: number[];
  labels: string[];
  color: string;
}

function UsageChart({ title, icon, values, labels, color }: UsageChartProps) {
  const max = Math.max(...values, 0);
  return (
    <Card
      title={
        <span className="flex items-center gap-2">
          {icon}
          {title}
        </span>
      }
    >
      <div
        className="grid gap-1 h-64 items-end"
        style={{
          gridTemplateColumns: `repeat(${values.length}, minmax(0, 1fr))`,
        }}
      >
        {values.map((value, index) => (
          <div key={labels[index]} className="flex h-full items-end">
            <div
              title={`${labels[index]}：${value} 次`}
              style={{
                backgroundColor: color,
                height: `${max ? Math.max((value / max) * 100, value ? 2 : 0) : 0}%`,
                width: "100%",
                minHeight: value ? 4 : 0,
                borderRadius: "4px 4px 0 0",
              }}
            />
          </div>
        ))}
      </div>
      <div className="flex justify-between">
        <Typography.Text type="secondary">{labels[0]}</Typography.Text>
        <Typography.Text type="secondary">{labels.at(-1)}</Typography.Text>
      </div>
    </Card>
  );
}

export default function UsagePage() {
  const [stats, setStats] = useState<UsageStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [today] = useState(() => Date.now());
  useEffect(() => {
    async function loadStats() {
      setIsLoading(true);
      const result = await api.usage_route_get();
      const data = responseData<UsageStats>(result);
      if (data) setStats(data);
      setIsLoading(false);
    }
    loadStats();
  }, []);

  if (isLoading) return <ConsolePageLoading />;
  if (!stats)
    return (
      <div className="flex justify-center py-12">
        <Typography.Text type="secondary">无法加载用量统计</Typography.Text>
      </div>
    );

  const todayLabels = Array.from({ length: 24 }, (_, hour) => `${hour}:00`);
  const days7Labels = Array.from({ length: 7 }, (_, index) =>
    new Date(today - (6 - index) * 86400000).toLocaleDateString("zh-CN", {
      weekday: "short",
    }),
  );
  const days30Labels = Array.from({ length: 30 }, (_, index) =>
    new Date(today - (29 - index) * 86400000).toLocaleDateString("zh-CN", {
      month: "short",
      day: "numeric",
    }),
  );
  const summaryCards = [
    {
      title: "今日消耗",
      value: stats.todayUsage,
      icon: <Activity size={20} />,
    },
    {
      title: "7 日消耗",
      value: stats.last7DaysUsage,
      icon: <Calendar size={20} />,
    },
    {
      title: "30 日消耗",
      value: stats.last30DaysUsage,
      icon: <TrendingUp size={20} />,
    },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div>
        <Typography.Title level={2}>用量统计</Typography.Title>
        <Typography.Text type="secondary">
          查看您的 API 使用情况
        </Typography.Text>
      </div>
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card>
            <Statistic
              title={
                <span className="flex items-center gap-2">
                  <Gift size={20} />
                  免费额度
                </span>
              }
              value={stats.freeCredits}
              suffix="积分"
            />
            <Typography.Text type="secondary">剩余可用积分</Typography.Text>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card>
            <Statistic
              title={
                <span className="flex items-center gap-2">
                  <Zap size={20} />
                  付费额度
                </span>
              }
              value={0}
              suffix="积分"
            />
            <Typography.Text type="secondary">已购买的额外积分</Typography.Text>
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]}>
        {summaryCards.map((card) => (
          <Col key={card.title} xs={24} md={8}>
            <Card>
              <Statistic
                title={card.title}
                value={card.value}
                prefix={card.icon}
              />
            </Card>
          </Col>
        ))}
      </Row>
      <Row gutter={[16, 16]}>

      <Col span={24}>
      <UsageChart
        title="今日消耗（按小时）"
        icon={<Clock size={20} />}
        values={stats.todayHourlyUsage}
        labels={todayLabels}
        color="#1677ff"
      />
      </Col>
      <Col span={24}>
      <UsageChart
        title="7 日消耗（按天）"
        icon={<Calendar size={20} />}
        values={stats.last7DaysDailyUsage}
        labels={days7Labels}
        color="#13c2c2"
      />
      </Col>
      <Col span={24}>
      <UsageChart
        title="30 日消耗（按天）"
        icon={<TrendingUp size={20} />}
        values={stats.last30DaysDailyUsage}
        labels={days30Labels}
        color="#722ed1"
      />
      </Col>
      </Row>
      <Card
        title={
          <span className="flex items-center gap-2">
            <BarChart3 size={20} />
            总计使用量
          </span>
        }
      >
        <Statistic value={stats.totalUsage} suffix="次" />
        <Progress percent={100} showInfo={false} />
      </Card>
    </div>
  );
}
