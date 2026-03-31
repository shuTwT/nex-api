export const paymentConfig = {
  wechat: {
    appId: process.env.WECHAT_PAY_APP_ID || '',
    mchId: process.env.WECHAT_PAY_MCH_ID || '',
    apiKey: process.env.WECHAT_PAY_API_KEY || '',
    privateKey: process.env.WECHAT_PAY_PRIVATE_KEY || '',
    publicKey: process.env.WECHAT_PAY_PUBLIC_KEY || '',
    paymentPublicKey: process.env.WECHAT_PAY_PAYMENT_PUBLIC_KEY || '',
    publicKeyId: process.env.WECHAT_PAY_PUBLIC_KEY_ID || '',
    notifyUrl: process.env.WECHAT_PAY_NOTIFY_URL || `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/callback/wechat`,
    debug: process.env.WECHAT_PAY_DEBUG === 'true',
  },
  alipay: {
    appId: process.env.ALIPAY_APP_ID || '',
    privateKey: process.env.ALIPAY_PRIVATE_KEY || '',
    alipayPublicKey: process.env.ALIPAY_PUBLIC_KEY || '',
    notifyUrl: process.env.ALIPAY_NOTIFY_URL || `${process.env.NEXT_PUBLIC_APP_URL}/api/payment/callback/alipay`,
    returnUrl: process.env.ALIPAY_RETURN_URL || `${process.env.NEXT_PUBLIC_APP_URL}/payment/result`,
    sandbox: process.env.ALIPAY_SANDBOX === 'true',
  },
  mock: {
    enabled: process.env.MOCK_PAYMENT_ENABLED === 'true',
    autoSuccess: process.env.MOCK_PAYMENT_AUTO_SUCCESS === 'true',
    delay: parseInt(process.env.MOCK_PAYMENT_DELAY || '2000'),
  },
};

export const isWechatPayConfigured = (): boolean => {
  const { appId, mchId, apiKey, privateKey, publicKey } = paymentConfig.wechat;
  return !!(appId && mchId && apiKey && privateKey && publicKey);
};

export const isAlipayConfigured = (): boolean => {
  const { appId, privateKey, alipayPublicKey } = paymentConfig.alipay;
  return !!(appId && privateKey && alipayPublicKey);
};

export const isMockPaymentEnabled = (): boolean => {
  return paymentConfig.mock.enabled;
};

export const getAvailablePaymentMethods = (): string[] => {
  const methods: string[] = [];
  
  if (isWechatPayConfigured()) {
    methods.push('wechat');
  }
  
  if (isAlipayConfigured()) {
    methods.push('alipay');
  }
  
  if (isMockPaymentEnabled()) {
    methods.push('mock');
  }
  
  return methods;
};
