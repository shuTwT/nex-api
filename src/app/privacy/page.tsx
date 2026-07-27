import type { Metadata } from "next";
import { MainLayout } from "@/components/main-layout";
import { Card, CardContent } from "@/components/ui/card";

export const metadata: Metadata = {
  title: "隐私政策 | NexApi",
  description: "NexApi 隐私政策",
};

const sections = [
  {
    title: "信息收集",
    content:
      "当您注册账号时，我们会收集您的邮箱地址、用户名和密码。当您使用平台服务时，我们会记录 API 调用日志、IP 地址、用户代理等信息，用于计费、审计和安全防护。我们不会收集与服务无关的个人信息。",
  },
  {
    title: "信息使用",
    content:
      "您的个人信息仅用于提供和改善平台服务，包括账号管理、积分计费、用量统计、安全审计和客户支持。我们不会将您的个人信息出售或出租给任何第三方，也不会用于与平台服务无关的商业目的。",
  },
  {
    title: "信息保护",
    content:
      "我们采用行业标准的安全措施保护您的个人信息，包括密码加盐哈希存储、HTTPS 传输加密、数据库访问控制等。密码在数据库中以 scrypt 哈希形式存储，我们无法也不会以明文形式查看您的密码。",
  },
  {
    title: "Cookie 使用",
    content:
      "本平台使用 Cookie 和类似技术维持登录会话状态、记录用户偏好（如主题设置、公告关闭状态）。您可以通过浏览器设置禁用 Cookie，但这可能影响部分功能的正常使用。我们不会使用 Cookie 进行跨站追踪。",
  },
  {
    title: "第三方服务",
    content:
      "本平台集成了 GitHub OAuth、统一身份认证（SSO）、支付宝和微信支付等第三方服务。当您使用这些服务时，需遵守相应第三方的隐私政策。第三方服务仅获取完成其功能所必需的信息，我们不会向其额外分享您的个人数据。",
  },
  {
    title: "用户权利",
    content:
      "您有权查阅、更正或删除您的个人信息，有权注销账号。如需行使上述权利，请通过平台提供的联系方式与我们联系。账号注销后，我们将在合理期限内删除您的个人数据，法律法规另有规定的除外。",
  },
];

export default function PrivacyPage() {
  return (
    <MainLayout>
      <section className="bg-gradient-to-br from-blue-50 via-white to-blue-50 py-12 md:py-16">
        <div className="container px-4 md:px-6 mx-auto">
          <div className="max-w-3xl mx-auto text-center">
            <h1 className="text-3xl md:text-4xl font-bold text-slate-900 mb-3">
              隐私政策
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
