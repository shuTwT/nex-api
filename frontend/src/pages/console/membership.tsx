import { useEffect, useState } from "react";
import { Badge, Button, Card, Col, Modal, Row, Tag, Typography } from "antd";
import { useNavigate } from "react-router";
import { useAuth } from "@/hooks/use-auth";
import { Crown, Check, Calendar, CreditCard, Zap, Loader2 } from "lucide-react";
import { api, responseData } from "@/lib/api";
import { ConsolePageLoading } from "@/components/console-page-loading";
import { toast } from "sonner";

interface SubscriptionPlan {
  id: string;
  title: string;
  price: number;
  totalCredits: number;
  sortOrder: number;
  validityDuration: number;
  validityUnit: string;
  creditResetCycle: string;
  isActive: boolean;
}
interface Subscription {
  id: string;
  planId: string | null;
  planName: string;
  credits: number;
  price: number;
  startDate: Date;
  endDate: Date;
  isActive: boolean;
  plan: SubscriptionPlan | null;
}

export default function MembershipPage() {
  const navigate = useNavigate();
  const { credits: userCredits } = useAuth();
  const [currentSubscription, setCurrentSubscription] =
    useState<Subscription | null>(null);
  const [availablePlans, setAvailablePlans] = useState<SubscriptionPlan[]>([]);
  const [paymentMethods, setPaymentMethods] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const [showPaymentDialog, setShowPaymentDialog] = useState(false);
  const [pendingPlan, setPendingPlan] = useState<SubscriptionPlan | null>(null);
  useEffect(() => {
    loadData();
  }, []);
  async function loadData() {
    setIsLoading(true);
    const [subscriptionResult, plansResult, paymentMethodsResult] =
      await Promise.all([
        api.membership_current_route_get(),
        api.membership_plans_route_get(),
        api.payment_methods_route_get(),
      ]);
    const subscriptionData = responseData<Subscription>(subscriptionResult);
    const plansData = responseData<SubscriptionPlan[]>(plansResult);
    const methodsData = responseData<string[]>(paymentMethodsResult);
    if (subscriptionData) setCurrentSubscription(subscriptionData);
    if (plansData) setAvailablePlans(plansData);
    if (methodsData) setPaymentMethods(methodsData);
    setIsLoading(false);
  }
  function handlePayment(plan: SubscriptionPlan) {
    if (plan.price === 0) return toast.error("免费计划无需支付");
    if (!paymentMethods.length)
      return toast.error("暂无可用支付方式，请先在系统设置中配置支付渠道");
    setPendingPlan(plan);
    setShowPaymentDialog(true);
  }
  async function handleConfirmPayment(method: string) {
    if (!pendingPlan) return;
    setIsProcessing(true);
    try {
      const result = await api.payment_methods_route_post({
        planId: pendingPlan.id,
        method,
      });
      const data = responseData<{ outTradeNo: string }>(result);
      if (result.success && data) {
        setShowPaymentDialog(false);
        toast.success("支付订单已创建");
        navigate(`/payment?outTradeNo=${data.outTradeNo}`);
      } else toast.error(result.error || "创建支付订单失败");
    } catch {
      toast.error("支付失败，请重试");
    } finally {
      setIsProcessing(false);
    }
  }
  const validityUnitLabels: Record<string, string> = {
    day: "天",
    week: "周",
    month: "月",
    year: "年",
  };
  const creditResetCycleLabels: Record<string, string> = {
    day: "每天",
    week: "每周",
    month: "每月",
    year: "每年",
  };
  const paymentMethodLabels: Record<string, { name: string; icon: string }> = {
    wechat: { name: "微信支付", icon: "💬" },
    alipay: { name: "支付宝", icon: "💳" },
    mock: { name: "模拟支付", icon: "🧪" },
  };
  if (isLoading) return <ConsolePageLoading />;
  return (
    <div className="space-y-6">
      <div>
        <Typography.Title level={2}>我的会员</Typography.Title>
        <Typography.Text type="secondary">管理您的订阅计划</Typography.Text>
      </div>
      <Card
        title={
          <span className="flex items-center gap-2">
            <Crown size={20} />
            当前订阅
          </span>
        }
        extra={
          <Tag color={currentSubscription?.isActive ? "success" : "default"}>
            {currentSubscription
              ? currentSubscription.isActive
                ? "有效"
                : "已过期"
              : "免费版"}
          </Tag>
        }
      >
        <Row gutter={[16, 16]}>
          <Col xs={24} md={6}>
            <Typography.Text type="secondary">计划名称</Typography.Text>
            <Typography.Paragraph>
              {currentSubscription?.planName || "免费版"}
            </Typography.Paragraph>
          </Col>
          <Col xs={24} md={6}>
            <Typography.Text type="secondary">剩余积分</Typography.Text>
            <Typography.Paragraph>
              {currentSubscription
                ? currentSubscription.credits.toLocaleString()
                : userCredits.toLocaleString()}
            </Typography.Paragraph>
          </Col>
          <Col xs={24} md={6}>
            <Typography.Text type="secondary">到期时间</Typography.Text>
            <Typography.Paragraph>
              {currentSubscription
                ? new Date(currentSubscription.endDate).toLocaleDateString(
                    "zh-CN",
                  )
                : "无限制"}
            </Typography.Paragraph>
          </Col>
          <Col xs={24} md={6}>
            <Typography.Text type="secondary">订阅价格</Typography.Text>
            <Typography.Paragraph>
              {currentSubscription ? `¥${currentSubscription.price}` : "¥0"}
            </Typography.Paragraph>
          </Col>
        </Row>
      </Card>
      <div>
        <Typography.Title level={3}>
          {currentSubscription ? "升级计划" : "选择计划"}
        </Typography.Title>
        <Row gutter={[16, 16]}>
          {availablePlans.map((plan) => {
            const isCurrentPlan = currentSubscription?.planId === plan.id;
            return (
              <Col key={plan.id} xs={24} md={12} xl={8}>
                <Card
                  title={plan.title}
                  extra={
                    isCurrentPlan && (
                      <Badge status="processing" text="当前计划" />
                    )
                  }
                >
                  <Typography.Title level={2}>
                    ¥{plan.price}
                    <Typography.Text type="secondary">
                      /{validityUnitLabels[plan.validityUnit]}
                    </Typography.Text>
                  </Typography.Title>
                  <Typography.Paragraph type="secondary">
                    {plan.totalCredits.toLocaleString()} 积分/
                    {creditResetCycleLabels[plan.creditResetCycle]}
                  </Typography.Paragraph>
                  {[
                    `${plan.totalCredits.toLocaleString()} 积分/${creditResetCycleLabels[plan.creditResetCycle]}`,
                    `有效期 ${plan.validityDuration} ${validityUnitLabels[plan.validityUnit]}`,
                    "全部接口访问",
                    "邮件+工单支持",
                  ].map((feature) => (
                    <div key={feature} className="flex items-center gap-2 py-1">
                      <Check size={16} color="#52c41a" />
                      {feature}
                    </div>
                  ))}
                  <Button
                    type="primary"
                    block
                    className="mt-6"
                    disabled={isCurrentPlan || isProcessing}
                    onClick={() => handlePayment(plan)}
                  >
                    {isProcessing ? (
                      <>
                        <Loader2
                          size={16}
                          className="mr-2 inline animate-spin"
                        />
                        处理中...
                      </>
                    ) : isCurrentPlan ? (
                      "当前计划"
                    ) : plan.price === 0 ? (
                      "免费订阅"
                    ) : (
                      "立即订阅"
                    )}
                  </Button>
                </Card>
              </Col>
            );
          })}
        </Row>
      </div>
      <Modal
        open={showPaymentDialog}
        title="选择支付方式"
        onCancel={() => setShowPaymentDialog(false)}
        footer={
          <Button
            onClick={() => setShowPaymentDialog(false)}
            disabled={isProcessing}
          >
            取消
          </Button>
        }
      >
        <Typography.Paragraph>
          {pendingPlan && (
            <>
              订阅 <Typography.Text strong>{pendingPlan.title}</Typography.Text>{" "}
              — ¥{pendingPlan.price}/
              {validityUnitLabels[pendingPlan.validityUnit]}
            </>
          )}
        </Typography.Paragraph>
        {paymentMethods.map((method) => {
          const methodInfo = paymentMethodLabels[method];
          if (!methodInfo) return null;
          return (
            <Button
              key={method}
              block
              className="mb-3"
              size="large"
              disabled={isProcessing}
              onClick={() => handleConfirmPayment(method)}
              icon={<CreditCard size={16} />}
            >
              {methodInfo.icon} {methodInfo.name}
            </Button>
          );
        })}
      </Modal>
    </div>
  );
}
