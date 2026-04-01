"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Check, X, Zap, Star, Crown, ArrowRight, HelpCircle } from "lucide-react";
import { MainLayout } from "@/components/main-layout";
import { BannerAd, InlineAd } from "@/components/ads";
import { AdPosition } from "@/types/ad-position";

const plans = [
  {
    name: "免费版",
    description: "适合个人开发者和小型项目",
    price: "0",
    period: "元/月",
    credits: "1000",
    color: "blue",
    icon: Zap,
    popular: false,
    features: [
      { name: "每月积分", value: "1000 积分", included: true },
      { name: "积分重置", value: "每月 1 号自动重置", included: true },
      { name: "API 调用", value: "基础接口访问", included: true },
      { name: "技术支持", value: "社区支持", included: true },
      { name: "文档访问", value: "完整文档", included: true },
      { name: "高级接口", value: "会员专享接口", included: false },
      { name: "优先支持", value: "优先工单处理", included: false },
      { name: "自定义集成", value: "专属集成服务", included: false },
      { name: "SLA 保障", value: "服务等级协议", included: false },
    ],
  },
  {
    name: "专业版",
    description: "适合成长型企业和专业开发者",
    price: "29.9",
    period: "元/月",
    credits: "3000",
    color: "cyan",
    icon: Star,
    popular: true,
    features: [
      { name: "每月积分", value: "3000 积分", included: true },
      { name: "积分重置", value: "每月 1 号自动重置", included: true },
      { name: "API 调用", value: "全部接口访问", included: true },
      { name: "技术支持", value: "邮件 + 工单支持", included: true },
      { name: "文档访问", value: "完整文档 + 示例", included: true },
      { name: "高级接口", value: "会员专享接口", included: true },
      { name: "优先支持", value: "优先工单处理", included: true },
      { name: "自定义集成", value: "专属集成服务", included: false },
      { name: "SLA 保障", value: "服务等级协议", included: false },
    ],
  },
  {
    name: "私有化部署",
    description: "适合大型企业和机构",
    price: "咨询",
    period: "销售",
    credits: "无限",
    color: "purple",
    icon: Crown,
    popular: false,
    features: [
      { name: "每月积分", value: "无限积分", included: true },
      { name: "积分重置", value: "无限制使用", included: true },
      { name: "API 调用", value: "全部接口 + 定制", included: true },
      { name: "技术支持", value: "7x24 专属支持", included: true },
      { name: "文档访问", value: "完整文档 + 培训", included: true },
      { name: "高级接口", value: "会员专享接口", included: true },
      { name: "优先支持", value: "专属客户经理", included: true },
      { name: "自定义集成", value: "专属集成服务", included: true },
      { name: "SLA 保障", value: "99.9% 可用性", included: true },
    ],
  },
];

export default function PricingPage() {
  return (
    <MainLayout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-cyan-50 via-white to-blue-50">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -left-40 w-80 h-80 bg-cyan-100/50 rounded-full blur-3xl"></div>
          <div className="absolute top-40 right-20 w-96 h-96 bg-blue-50/50 rounded-full blur-3xl"></div>
          <div className="absolute bottom-20 left-1/3 w-72 h-72 bg-indigo-50/50 rounded-full blur-3xl"></div>
        </div>

        <div className="container relative px-4 py-16 md:py-24 md:px-6 mx-auto">
          <div className="flex flex-col items-center text-center space-y-6">
            <div className="space-y-4 max-w-3xl">
              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight text-slate-900">
                简单透明的定价
              </h1>
              <p className="text-lg md:text-xl text-slate-600">
                选择适合您的方案，随时升级，无隐藏费用
              </p>
            </div>

            {/* Trust Badges */}
            <div className="flex flex-wrap items-center justify-center gap-6 pt-4">
              <div className="flex items-center gap-2 text-sm text-slate-600">
                <Check className="h-4 w-4 text-green-600" />
                <span>无需信用卡</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-slate-600">
                <Check className="h-4 w-4 text-green-600" />
                <span>随时取消</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-slate-600">
                <Check className="h-4 w-4 text-green-600" />
                <span>7 天无理由退款</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Pricing Cards */}
      <section className="container px-4 py-16 md:px-6 mx-auto">
        <div className="grid gap-8 md:grid-cols-3">
          {plans.map((plan) => (
            <Card
              key={plan.name}
              className={`relative overflow-visible flex flex-col h-full ${
                plan.popular
                  ? "border-2 border-cyan-500 shadow-xl scale-105"
                  : "border border-slate-200 shadow-md"
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 -translate-x-1/2">
                  <Badge className="bg-gradient-to-r from-cyan-500 to-blue-500 border-0 px-4 py-1">
                    最受欢迎
                  </Badge>
                </div>
              )}

              <CardHeader className="pb-4">
                <div className="flex items-center gap-3 mb-4">
                  <div className={`p-2 rounded-lg ${
                    plan.color === "blue" ? "bg-blue-100" :
                    plan.color === "cyan" ? "bg-cyan-100" : "bg-purple-100"
                  }`}>
                    <plan.icon className={`h-6 w-6 ${
                      plan.color === "blue" ? "text-blue-600" :
                      plan.color === "cyan" ? "text-cyan-600" : "text-purple-600"
                    }`} />
                  </div>
                  <div>
                    <h3 className="text-xl font-bold text-slate-900">{plan.name}</h3>
                    <p className="text-sm text-slate-500">{plan.description}</p>
                  </div>
                </div>

                <div className="space-y-1">
                  <div className="flex items-baseline gap-1">
                    <span className="text-4xl font-bold text-slate-900">¥{plan.price}</span>
                    <span className="text-slate-500">/{plan.period}</span>
                  </div>
                  <div className="text-sm text-slate-600">
                    每月 <span className="font-semibold text-slate-900">{plan.credits} 积分</span>
                    <span className="text-slate-500 text-xs ml-1">(每月 1 号重置)</span>
                  </div>
                </div>
              </CardHeader>

              <CardContent className="flex-1">
                <div className="space-y-3">
                  {plan.features.map((feature) => (
                    <div key={feature.name} className="flex items-start gap-3">
                      <div className="flex-shrink-0 mt-0.5">
                        {feature.included ? (
                          <Check className="h-5 w-5 text-green-600" />
                        ) : (
                          <X className="h-5 w-5 text-slate-300" />
                        )}
                      </div>
                      <div className="flex-1">
                        <div className="text-sm text-slate-700">{feature.name}</div>
                        <div className="text-xs text-slate-500">{feature.value}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>

              <CardFooter className="pt-4">
                <Button
                  className={`w-full cursor-pointer ${
                    plan.popular
                      ? "bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-600 hover:to-blue-600"
                      : plan.name === "私有化部署"
                      ? "bg-purple-600 hover:bg-purple-700"
                      : "bg-blue-600 hover:bg-blue-700"
                  }`}
                  size="lg"
                >
                  {plan.name === "免费版" ? "免费开始" :
                   plan.name === "私有化部署" ? "联系销售" :
                   "立即订阅"}
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>
      </section>

      {/* FAQ Section */}
      <section className="bg-slate-50 dark:bg-slate-800 py-16">
        <div className="container px-4 md:px-6 mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold tracking-tight mb-4 text-slate-900 dark:text-white">
              常见问题
            </h2>
            <p className="text-slate-600 max-w-2xl mx-auto">
              了解更多关于定价和积分的问题
            </p>
          </div>

          <div className="grid gap-6 md:grid-cols-2 max-w-4xl mx-auto">
            <Card className="border-0 shadow-md">
              <CardContent className="p-6 space-y-3">
                <div className="flex items-start gap-3">
                  <HelpCircle className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-slate-900 dark:text-white">积分如何重置？</h3>
                    <p className="text-sm text-slate-600 mt-1">
                      所有计划的积分都会在每月 1 号 0 点自动重置。未使用的积分不会累积到下个月。
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-md">
              <CardContent className="p-6 space-y-3">
                <div className="flex items-start gap-3">
                  <HelpCircle className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-slate-900 dark:text-white">如何计算 API 调用消耗的积分？</h3>
                    <p className="text-sm text-slate-600 mt-1">
                      不同 API 接口消耗的积分不同。基础接口通常消耗 1 积分/次，高级接口可能消耗 5-10 积分/次。
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-md">
              <CardContent className="p-6 space-y-3">
                <div className="flex items-start gap-3">
                  <HelpCircle className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-slate-900 dark:text-white">可以随时升级或降级吗？</h3>
                    <p className="text-sm text-slate-600 mt-1">
                      是的，您可以随时升级或降级您的计划。升级会立即生效，降级会在下个计费周期生效。
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-md">
              <CardContent className="p-6 space-y-3">
                <div className="flex items-start gap-3">
                  <HelpCircle className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-slate-900 dark:text-white">私有化部署包含哪些服务？</h3>
                    <p className="text-sm text-slate-600 mt-1">
                      私有化部署包含完整的 API 系统部署到您的服务器、专属技术支持、定制开发、SLA 保障等服务。
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Inline Ad in FAQ */}
            <div className="md:col-span-2">
              <InlineAd size="sm" position={AdPosition.CONTENT_BOTTOM} />
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container px-4 py-16 md:px-6 mx-auto">
        <Card className="bg-gradient-to-r from-blue-600 to-cyan-600 border-0">
          <CardContent className="p-12 text-center text-white">
            <h2 className="text-3xl font-bold mb-4">
              准备好开始了吗？
            </h2>
            <p className="text-lg text-white/90 mb-8 max-w-2xl mx-auto">
              立即注册，免费开始使用。无需信用卡，随时可以取消。
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button size="lg" className="bg-white text-blue-600 hover:bg-white/90 cursor-pointer">
                免费注册
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              <Button size="lg" variant="outline" className="border-2 border-white text-white hover:bg-white/10 bg-transparent cursor-pointer">
                联系销售
              </Button>
            </div>
          </CardContent>
        </Card>
      </section>
    </MainLayout>
  );
}
