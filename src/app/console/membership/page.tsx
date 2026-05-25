"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Crown, Check, Calendar, CreditCard, Zap, Loader2 } from "lucide-react";
import { api } from "@/lib/api-client";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

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
  const router = useRouter();
  const [currentSubscription, setCurrentSubscription] = useState<Subscription | null>(null);
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
    const [subscriptionResult, plansResult, paymentMethodsResult] = await Promise.all([
      api.get("/api/membership/current"),
      api.get("/api/membership/plans"),
      api.get("/api/payment/methods"),
    ]);

    if (subscriptionResult.success && subscriptionResult.data) {
      setCurrentSubscription(subscriptionResult.data);
    }

    if (plansResult.success && plansResult.data) {
      setAvailablePlans(plansResult.data);
    }

    if (paymentMethodsResult.success && paymentMethodsResult.data) {
      setPaymentMethods(paymentMethodsResult.data);
    }

    setIsLoading(false);
  }

  function handlePayment(plan: SubscriptionPlan) {
    if (plan.price === 0) {
      toast.error("免费计划无需支付");
      return;
    }

    if (paymentMethods.length === 0) {
      toast.error("暂无可用支付方式，请先在系统设置中配置支付渠道");
      return;
    }

    setPendingPlan(plan);
    setShowPaymentDialog(true);
  }

  async function handleConfirmPayment(method: string) {
    if (!pendingPlan) return;

    setIsProcessing(true);

    try {
      const result = await api.post("/api/payment/methods", { planId: pendingPlan.id, method });

      if (result.success && result.data) {
        setShowPaymentDialog(false);
        toast.success("支付订单已创建");
        router.push(`/payment?outTradeNo=${result.data.outTradeNo}`);
      } else {
        toast.error(result.error || "创建支付订单失败");
      }
    } catch (error) {
      console.error("支付失败:", error);
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

  const paymentMethodLabels: Record<string, { name: string; icon: string; color: string }> = {
    wechat: { name: "微信支付", icon: "💬", color: "bg-green-500" },
    alipay: { name: "支付宝", icon: "💳", color: "bg-blue-500" },
    mock: { name: "模拟支付", icon: "🧪", color: "bg-purple-500" },
  };

  const { credits: userCredits } = useAuth();

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
        <h1 className="text-2xl font-bold text-slate-900">我的会员</h1>
        <p className="text-slate-500 mt-1">管理您的订阅计划</p>
      </div>

      <Card className={currentSubscription ? "border-blue-200 bg-gradient-to-r from-blue-50 to-cyan-50" : "border-slate-200"}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Crown className={`h-5 w-5 ${currentSubscription ? "text-blue-600" : "text-slate-400"}`} />
              当前订阅
            </CardTitle>
            <Badge className={currentSubscription?.isActive ? "bg-green-100 text-green-700 border-green-200" : "bg-slate-100 text-slate-600 border-slate-200"}>
              {currentSubscription ? (currentSubscription.isActive ? "有效" : "已过期") : "免费版"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center">
                <Crown className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-slate-500">计划名称</p>
                <p className="font-medium text-slate-900">
                  {currentSubscription ? currentSubscription.planName : "免费版"}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-green-100 flex items-center justify-center">
                <Zap className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-slate-500">剩余积分</p>
                <p className="font-medium text-slate-900">
                  {currentSubscription ? currentSubscription.credits.toLocaleString() : userCredits.toLocaleString()}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-purple-100 flex items-center justify-center">
                <Calendar className="h-5 w-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-slate-500">到期时间</p>
                <p className="font-medium text-slate-900">
                  {currentSubscription
                    ? new Date(currentSubscription.endDate).toLocaleDateString("zh-CN")
                    : "无限制"}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-orange-100 flex items-center justify-center">
                <CreditCard className="h-5 w-5 text-orange-600" />
              </div>
              <div>
                <p className="text-sm text-slate-500">订阅价格</p>
                <p className="font-medium text-slate-900">
                  {currentSubscription ? `¥${currentSubscription.price}` : "¥0"}
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <div>
        <h2 className="text-lg font-semibold text-slate-900 mb-4">
          {currentSubscription ? "升级计划" : "选择计划"}
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {availablePlans.map((plan) => {
            const isCurrentPlan = currentSubscription?.planId === plan.id;

            return (
              <Card
                key={plan.id}
                className={`hover:shadow-md transition-shadow ${
                  isCurrentPlan ? "border-blue-500 border-2" : ""
                }`}
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{plan.title}</CardTitle>
                    {isCurrentPlan && (
                      <Badge className="bg-blue-100 text-blue-700 border-blue-200">
                        当前计划
                      </Badge>
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="mb-4">
                    <div className="flex items-baseline gap-1">
                      <span className="text-3xl font-bold text-slate-900">
                        ¥{plan.price}
                      </span>
                      <span className="text-slate-500">
                        /{validityUnitLabels[plan.validityUnit]}
                      </span>
                    </div>
                    <p className="text-sm text-slate-500 mt-1">
                      {plan.totalCredits.toLocaleString()} 积分/{creditResetCycleLabels[plan.creditResetCycle]}
                    </p>
                  </div>

                  <ul className="space-y-3 mb-6">
                    <li className="flex items-center gap-2 text-sm text-slate-600">
                      <Check className="h-4 w-4 text-green-600 flex-shrink-0" />
                      <span>{plan.totalCredits.toLocaleString()} 积分/{creditResetCycleLabels[plan.creditResetCycle]}</span>
                    </li>
                    <li className="flex items-center gap-2 text-sm text-slate-600">
                      <Check className="h-4 w-4 text-green-600 flex-shrink-0" />
                      <span>有效期 {plan.validityDuration} {validityUnitLabels[plan.validityUnit]}</span>
                    </li>
                    <li className="flex items-center gap-2 text-sm text-slate-600">
                      <Check className="h-4 w-4 text-green-600 flex-shrink-0" />
                      <span>全部接口访问</span>
                    </li>
                    <li className="flex items-center gap-2 text-sm text-slate-600">
                      <Check className="h-4 w-4 text-green-600 flex-shrink-0" />
                      <span>邮件+工单支持</span>
                    </li>
                  </ul>

                  <Button
                    className="w-full cursor-pointer"
                    variant={isCurrentPlan ? "outline" : "default"}
                    disabled={isCurrentPlan || isProcessing}
                    onClick={() => handlePayment(plan)}
                  >
                    {isProcessing ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>

      <Dialog open={showPaymentDialog} onOpenChange={setShowPaymentDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>选择支付方式</DialogTitle>
            <DialogDescription>
              {pendingPlan && (
                <>
                  订阅 <span className="font-medium text-slate-900">{pendingPlan.title}</span>
                  {" "}— ¥{pendingPlan.price}/{validityUnitLabels[pendingPlan.validityUnit]}
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-3 py-4">
            {paymentMethods.map((method) => {
              const methodInfo = paymentMethodLabels[method];
              return (
                <button
                  key={method}
                  onClick={() => handleConfirmPayment(method)}
                  disabled={isProcessing}
                  className="flex items-center gap-4 p-4 rounded-lg border-2 border-slate-200 hover:border-blue-400 hover:bg-blue-50 transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <div
                    className={`h-10 w-10 rounded-lg ${methodInfo.color} flex items-center justify-center text-white text-xl flex-shrink-0`}
                  >
                    {methodInfo.icon}
                  </div>
                  <div className="text-left flex-1">
                    <p className="font-medium text-slate-900">{methodInfo.name}</p>
                  </div>
                  {isProcessing ? (
                    <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
                  ) : (
                    <CreditCard className="h-4 w-4 text-slate-400" />
                  )}
                </button>
              );
            })}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowPaymentDialog(false)} disabled={isProcessing}>
              取消
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
