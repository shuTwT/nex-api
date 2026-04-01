import { getConfigByCategory } from "../config";

export interface PaymentConfig {
  wechat: {
    appId: string;
    mchId: string;
    apiKey: string;
    privateKey: string;
    publicKey: string;
    paymentPublicKey: string;
    publicKeyId: string;
    notifyUrl: string;
    debug: boolean;
  };
  alipay: {
    appId: string;
    privateKey: string;
    alipayPublicKey: string;
    notifyUrl: string;
    returnUrl: string;
    sandbox: boolean;
  };
  mock: {
    enabled: boolean;
    autoSuccess: boolean;
    delay: number;
  };
}

let cachedConfig: PaymentConfig | null = null;
let cacheTime: number = 0;
const CACHE_TTL = 60000;

export async function getPaymentConfig(): Promise<PaymentConfig> {
  const now = Date.now();
  
  if (cachedConfig && (now - cacheTime) < CACHE_TTL) {
    return cachedConfig;
  }

  const settingsMap: Record<string, string> = await getConfigByCategory("payment");

  const appUrl = process.env.NEXT_PUBLIC_APP_URL || "";

  cachedConfig = {
    wechat: {
      appId: settingsMap.wechatPayAppId || "",
      mchId: settingsMap.wechatPayMchId || "",
      apiKey: settingsMap.wechatPayApiKey || "",
      privateKey: settingsMap.wechatPayPrivateKey || "",
      publicKey: settingsMap.wechatPayPublicKey || "",
      paymentPublicKey: settingsMap.wechatPayPaymentPublicKey || "",
      publicKeyId: settingsMap.wechatPayPublicKeyId || "",
      notifyUrl: settingsMap.wechatPayNotifyUrl || `${appUrl}/api/payment/callback/wechat`,
      debug: settingsMap.wechatPayDebug === "true",
    },
    alipay: {
      appId: settingsMap.alipayAppId || "",
      privateKey: settingsMap.alipayPrivateKey || "",
      alipayPublicKey: settingsMap.alipayPublicKey || "",
      notifyUrl: settingsMap.alipayNotifyUrl || `${appUrl}/api/payment/callback/alipay`,
      returnUrl: settingsMap.alipayReturnUrl || `${appUrl}/payment/result`,
      sandbox: settingsMap.alipaySandbox === "true",
    },
    mock: {
      enabled: settingsMap.mockPaymentEnabled === "true",
      autoSuccess: settingsMap.mockPaymentAutoSuccess === "true",
      delay: parseInt(settingsMap.mockPaymentDelay || "2000"),
    },
  };

  cacheTime = now;
  
  return cachedConfig;
}

export async function isWechatPayConfigured(): Promise<boolean> {
  const config = await getPaymentConfig();
  const { appId, mchId, apiKey, privateKey, publicKey } = config.wechat;
  return !!(appId && mchId && apiKey && privateKey && publicKey);
}

export async function isAlipayConfigured(): Promise<boolean> {
  const config = await getPaymentConfig();
  const { appId, privateKey, alipayPublicKey } = config.alipay;
  return !!(appId && privateKey && alipayPublicKey);
}

export async function isMockPaymentEnabled(): Promise<boolean> {
  const config = await getPaymentConfig();
  return config.mock.enabled;
}

export async function getAvailablePaymentMethods(): Promise<string[]> {
  const methods: string[] = [];
  
  if (await isWechatPayConfigured()) {
    methods.push("wechat");
  }
  
  if (await isAlipayConfigured()) {
    methods.push("alipay");
  }
  
  if (await isMockPaymentEnabled()) {
    methods.push("mock");
  }
  
  return methods;
}

export function clearPaymentConfigCache(): void {
  cachedConfig = null;
  cacheTime = 0;
}
