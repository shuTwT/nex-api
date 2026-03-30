"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Crown, Check, Calendar, CreditCard, Zap } from "lucide-react";
import { getCurrentSubscription, getAvailablePlans, subscribeToPlan } from "@/app/actions/membership";
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
  const [currentSubscription, setCurrentSubscription] = useState<Subscription | null>(null);
  const [availablePlans, setAvailablePlans] = useState<SubscriptionPlan[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [subscribingPlanId, setSubscribingPlanId] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setIsLoading(true);
    const [subscriptionResult, plansResult] = await Promise.all([
      getCurrentSubscription(),
      getAvailablePlans(),
    ]);

    if (subscriptionResult.success && subscriptionResult.data) {
      setCurrentSubscription(subscriptionResult.data);
    }

    if (plansResult.success && plansResult.data) {
      setAvailablePlans(plansResult.data);
    }

    setIsLoading(false);
  }

  async function handleSubscribe(planId: string) {
    setSubscribingPlanId(planId);
    const result = await subscribeToPlan(planId);

    if (result.success) {
      toast.success("订阅成功！");
      await loadData();
    } else {
      toast.error(result.error || "订阅失败");
    }

    setSubscribingPlanId(null);
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

      {currentSubscription && (
        <Card className="border-blue-200 bg-gradient-to-r from-blue-50 to-cyan-50">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <Crown className="h-5 w-5 text-blue-600" />
                当前订阅
              </CardTitle>
              <Badge className="bg-green-100 text-green-700 border-green-200">
                {currentSubscription.isActive ? "有效" : "已过期"}
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
                  <p className="font-medium text-slate-900">{currentSubscription.planName}</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-lg bg-green-100 flex items-center justify-center">
                  <Zap className="h-5 w-5 text-green-600" />
                </div>
                <div>
                  <p className="text-sm text-slate-500">剩余积分</p>
                  <p className="font-medium text-slate-900">{currentSubscription.credits.toLocaleString()}</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-lg bg-purple-100 flex items-center justify-center">
                  <Calendar className="h-5 w-5 text-purple-600" />
                </div>
                <div>
                  <p className="text-sm text-slate-500">到期时间</p>
                  <p className="font-medium text-slate-900">
                    {new Date(currentSubscription.endDate).toLocaleDateString("zh-CN")}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="h-10 w-10 rounded-lg bg-orange-100 flex items-center justify-center">
                  <CreditCard className="h-5 w-5 text-orange-600" />
                </div>
                <div>
                  <p className="text-sm text-slate-500">订阅价格</p>
                  <p className="font-medium text-slate-900">¥{currentSubscription.price}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div>
        <h2 className="text-lg font-semibold text-slate-900 mb-4">
          {currentSubscription ? "升级计划" : "选择计划"}
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {availablePlans.map((plan) => {
            const isCurrentPlan = currentSubscription?.planId === plan.id;
            const isSubscribing = subscribingPlanId === plan.id;

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
                    disabled={isCurrentPlan || isSubscribing}
                    onClick={() => handleSubscribe(plan.id)}
                  >
                    {isSubscribing ? "订阅中..." : isCurrentPlan ? "当前计划" : plan.price === 0 ? "免费订阅" : "立即订阅"}
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </div>
  );
}
