"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Crown, Check } from "lucide-react";

const currentPlan = {
  name: "专业版",
  price: "29.9",
  period: "月",
  credits: "3000",
  usedCredits: 1234,
  startDate: "2024-01-01",
  endDate: "2024-12-31",
  status: "active",
};

const plans = [
  {
    name: "免费版",
    price: "0",
    period: "月",
    credits: "1000",
    features: ["1000 积分/月", "基础接口访问", "社区支持"],
    current: false,
    color: "gray",
  },
  {
    name: "专业版",
    price: "29.9",
    period: "月",
    credits: "3000",
    features: ["3000 积分/月", "全部接口访问", "邮件+工单支持", "会员专享接口"],
    current: true,
    color: "blue",
    popular: true,
  },
  {
    name: "企业版",
    price: "联系我们",
    period: "",
    credits: "无限",
    features: ["无限积分", "专属客服", "定制化服务", "SLA 保障"],
    current: false,
    color: "purple",
  },
];

export default function MembershipPage() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">我的会员</h1>
        <p className="text-slate-500 mt-1">管理您的订阅计划和积分使用</p>
      </div>

      {/* Current Plan */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">当前计划</CardTitle>
            <Badge className="bg-green-50 text-green-700 border-green-200">
              {currentPlan.status === "active" ? "已激活" : "已过期"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
            <div>
              <p className="text-sm text-slate-500">计划名称</p>
              <div className="flex items-center gap-2 mt-1">
                <Crown className="h-5 w-5 text-blue-600" />
                <span className="text-xl font-bold text-slate-900">{currentPlan.name}</span>
              </div>
            </div>
            <div>
              <p className="text-sm text-slate-500">月费</p>
              <p className="text-xl font-bold text-slate-900 mt-1">¥{currentPlan.price}/月</p>
            </div>
            <div>
              <p className="text-sm text-slate-500">积分使用</p>
              <div className="mt-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xl font-bold text-slate-900">
                    {currentPlan.usedCredits.toLocaleString()}
                  </span>
                  <span className="text-slate-500">/</span>
                  <span className="text-slate-600">{currentPlan.credits.toLocaleString()}</span>
                </div>
                <div className="w-full bg-slate-100 rounded-full h-2">
                  <div 
                    className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${(currentPlan.usedCredits / parseInt(currentPlan.credits)) * 100}%` }}
                  />
                </div>
              </div>
            </div>
            <div>
              <p className="text-sm text-slate-500">有效期至</p>
              <p className="text-xl font-bold text-slate-900 mt-1">{currentPlan.endDate}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Plan Comparison */}
      <div>
        <h2 className="text-lg font-bold text-slate-900 mb-4">升级计划</h2>
        <div className="grid gap-6 md:grid-cols-3">
          {plans.map((plan) => {
            const colorClasses = {
              gray: "border-slate-200",
              blue: "border-blue-500 ring-2 ring-blue-500",
              purple: "border-purple-200",
            };
            
            return (
              <Card 
                key={plan.name} 
                className={`relative overflow-visible ${colorClasses[plan.color as keyof typeof colorClasses]} ${
                  plan.current ? "" : "hover:shadow-lg transition-shadow cursor-pointer"
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                    <Badge className="bg-blue-600 text-white border-0">最受欢迎</Badge>
                  </div>
                )}
                <CardContent className="p-6">
                  <div className="text-center mb-6">
                    <h3 className="text-xl font-bold text-slate-900">{plan.name}</h3>
                    <div className="mt-3">
                      <span className="text-3xl font-bold text-slate-900">¥{plan.price}</span>
                      {plan.period && (
                        <span className="text-slate-500">/{plan.period}</span>
                      )}
                    </div>
                    <p className="text-sm text-slate-500 mt-1">
                      {plan.credits === "无限" ? "无限积分" : `${plan.credits} 积分/月`}
                    </p>
                  </div>
                  
                  <ul className="space-y-3 mb-6">
                    {plan.features.map((feature, index) => (
                      <li key={index} className="flex items-center gap-2 text-sm text-slate-600">
                        <Check className="h-4 w-4 text-green-600 flex-shrink-0" />
                        <span>{feature}</span>
                      </li>
                    ))}
                  </ul>
                  
                  <Button 
                    className="w-full cursor-pointer"
                    variant={plan.current ? "outline" : "default"}
                    disabled={plan.current}
                  >
                    {plan.current ? "当前计划" : "升级"}
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
