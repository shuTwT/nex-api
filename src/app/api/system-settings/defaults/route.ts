import { NextRequest, NextResponse } from "next/server";
import { apiSuccess } from "@/lib/api-auth";

export async function GET(_request: NextRequest) {
  return apiSuccess({
    general: [
      { key: "siteName", value: "API 网关", category: "general", description: "网站名称" },
      { key: "siteDescription", value: "一站式 API 服务平台", category: "general", description: "网站描述" },
      { key: "siteLogo", value: "", category: "general", description: "网站 Logo" },
      { key: "contactEmail", value: "support@example.com", category: "general", description: "联系邮箱" },
    ],
    operation: {
      basic: [
        { key: "registrationEnabled", value: "true", category: "operation", description: "是否允许用户注册" },
        { key: "defaultCredits", value: "1000", category: "operation", description: "新用户默认积分" },
        { key: "inviteRewards", value: "100", category: "operation", description: "邀请奖励积分" },
        { key: "maintenanceMode", value: "false", category: "operation", description: "是否开启维护模式" },
      ],
      announcement: [
        { key: "announcementEnabled", value: "false", category: "operation", description: "是否启用公告" },
        { key: "announcementContent", value: "", category: "operation", description: "公告内容" },
      ],
    },
    payment: {
      basic: [
        { key: "alipayEnabled", value: "false", category: "payment", description: "是否开启支付宝支付" },
        { key: "wechatEnabled", value: "false", category: "payment", description: "是否开启微信支付" },
        { key: "creditPrice", value: "1", category: "payment", description: "每积分价格（元）" },
        { key: "minRecharge", value: "10", category: "payment", description: "最低充值金额（元）" },
        { key: "mockPaymentEnabled", value: "false", category: "payment", description: "是否启用模拟支付" },
        { key: "mockPaymentAutoSuccess", value: "true", category: "payment", description: "模拟支付自动成功" },
        { key: "mockPaymentDelay", value: "2000", category: "payment", description: "模拟支付延迟时间（毫秒）" },
      ],
      alipay: [
        { key: "alipayAppId", value: "", category: "payment", description: "支付宝 AppID" },
        { key: "alipayPrivateKey", value: "", category: "payment", description: "支付宝私钥" },
        { key: "alipayPublicKey", value: "", category: "payment", description: "支付宝公钥" },
        { key: "alipayNotifyUrl", value: "", category: "payment", description: "支付宝回调地址" },
        { key: "alipayReturnUrl", value: "", category: "payment", description: "支付宝返回地址" },
        { key: "alipaySandbox", value: "false", category: "payment", description: "支付宝沙箱模式" },
      ],
      wechat: [
        { key: "wechatPayAppId", value: "", category: "payment", description: "微信支付 AppID" },
        { key: "wechatPayMchId", value: "", category: "payment", description: "微信支付商户号" },
        { key: "wechatPayApiKey", value: "", category: "payment", description: "微信支付 API 密钥" },
        { key: "wechatPayPrivateKey", value: "", category: "payment", description: "微信支付私钥" },
        { key: "wechatPayPublicKey", value: "", category: "payment", description: "微信支付公钥" },
        { key: "wechatPayPaymentPublicKey", value: "", category: "payment", description: "微信支付平台公钥" },
        { key: "wechatPayPublicKeyId", value: "", category: "payment", description: "微信支付公钥 ID" },
        { key: "wechatPayNotifyUrl", value: "", category: "payment", description: "微信支付回调地址" },
        { key: "wechatPayDebug", value: "false", category: "payment", description: "微信支付调试模式" },
      ],
    },
    oauth: {
      basic: [
        { key: "oauthProviders", value: "[]", category: "oauth", description: "OAuth 提供商配置" },
      ],
    },
  });
}
