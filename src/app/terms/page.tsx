import type { Metadata } from "next";
import { MainLayout } from "@/components/main-layout";
import { Card, CardContent } from "@/components/ui/card";

export const metadata: Metadata = {
  title: "服务条款 | NexApi",
  description: "NexApi 服务条款",
};

const sections = [
  {
    title: "服务说明",
    content:
      "NexApi 是一站式 API 聚合管理平台，为开发者提供 HTTP API 接口管理、MCP 服务管理、API 市场、积分充值与订阅等服务。使用本平台即表示您同意遵守本服务条款。",
  },
  {
    title: "账号与安全",
    content:
      "您需注册账号后方可使用本平台的部分功能。您应妥善保管账号和 API Token，因账号泄露导致的损失由您自行承担。您不得将账号转借、出售给他人使用。如发现账号被盗用，请立即联系平台管理员。",
  },
  {
    title: "API 使用规范",
    content:
      "您应按照接口文档规范调用 API，不得利用本平台接口从事违法活动，不得进行恶意刷量、爬虫攻击、DDoS 等行为。平台有权对异常调用行为进行限流或封禁处理，且不退还已消耗的积分。",
  },
  {
    title: "付费与积分",
    content:
      "积分可通过充值、兑换码或订阅计划获取。积分一经消耗不予退还，充值支付成功后不支持退款（法律另有规定的除外）。订阅计划按周期计费，可随时取消，取消后当前周期内仍可继续使用。",
  },
  {
    title: "免责声明",
    content:
      "本平台提供的第三方 API 接口由上游服务商提供，我们不对上游接口的可用性、准确性和稳定性做担保。因上游服务中断或数据错误导致的损失，本平台不承担赔偿责任。",
  },
  {
    title: "协议变更",
    content:
      "本平台保留随时修改本服务条款的权利。条款变更后将在平台上公告，您继续使用本平台即视为同意变更后的条款。如您不同意变更后的条款，请停止使用本平台。",
  },
];

export default function TermsPage() {
  return (
    <MainLayout>
      <section className="bg-gradient-to-br from-blue-50 via-white to-blue-50 py-12 md:py-16">
        <div className="container px-4 md:px-6 mx-auto">
          <div className="max-w-3xl mx-auto text-center">
            <h1 className="text-3xl md:text-4xl font-bold text-slate-900 mb-3">
              服务条款
            </h1>
            <p className="text-slate-600">
              最后更新日期：2026 年 7 月
            </p>
          </div>
        </div>
      </section>

      <section className="container px-4 py-12 md:px-6 mx-auto">
        <div className="max-w-3xl mx-auto space-y-6">
          {sections.map((section, index) => (
            <Card key={section.title} className="border border-slate-200 shadow-sm">
              <CardContent className="p-6">
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-sm font-semibold text-blue-600">
                    {index + 1}
                  </div>
                  <div className="flex-1">
                    <h2 className="text-lg font-semibold text-slate-900 mb-2">
                      {section.title}
                    </h2>
                    <p className="text-sm text-slate-600 leading-relaxed">
                      {section.content}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>
    </MainLayout>
  );
}
