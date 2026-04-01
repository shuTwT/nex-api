import WeChatPay from 'better-wechatpay';
import { getPaymentConfig, isWechatPayConfigured } from './config';
import type { PaymentService, CreatePaymentParams, CreatePaymentResult, PaymentCallbackData } from './types';
import { createPaymentRecord, updatePaymentStatus, getPaymentByOutTradeNo } from './utils';

class WechatPaymentService implements PaymentService {
  private wechat: WeChatPay | null = null;
  private initPromise: Promise<void> | null = null;

  private async ensureInitialized(): Promise<void> {
    if (this.wechat) return;
    
    if (!this.initPromise) {
      this.initPromise = this.initialize();
    }
    
    await this.initPromise;
  }

  private async initialize(): Promise<void> {
    if (!(await isWechatPayConfigured())) {
      return;
    }

    const config = await getPaymentConfig();
    this.wechat = new WeChatPay({
      config: {
        appId: config.wechat.appId,
        mchId: config.wechat.mchId,
        apiKey: config.wechat.apiKey,
        privateKey: config.wechat.privateKey,
        publicKey: config.wechat.publicKey,
        notifyUrl: config.wechat.notifyUrl,
        debug: config.wechat.debug,
        paymentPublicKey: config.wechat.paymentPublicKey || undefined,
        publicKeyId: config.wechat.publicKeyId || undefined,
      },
    });
  }

  async createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult> {
    await this.ensureInitialized();
    
    if (!this.wechat) {
      return { success: false, error: '微信支付未配置' };
    }

    try {
      const outTradeNo = `WX${Date.now()}${Math.random().toString(36).substring(2, 8)}`;
      
      const result = await this.wechat.native.create({
        out_trade_no: outTradeNo,
        description: '支付订单',
        amount: Math.round(params.amount * 100),
      });

      if (!result.code_url) {
        return { success: false, error: '创建支付二维码失败' };
      }

      const expiredAt = new Date();
      expiredAt.setHours(expiredAt.getHours() + 2);

      const payment = await createPaymentRecord({
        userId: params.userId,
        outTradeNo,
        method: 'wechat',
        amount: params.amount,
        qrcodeUrl: result.code_url,
        notifyUrl: params.notifyUrl,
        expiredAt,
        metadata: params.metadata,
      });

      return {
        success: true,
        paymentId: payment.id,
        outTradeNo,
        qrcodeUrl: result.code_url,
      };
    } catch (error) {
      console.error('创建微信支付失败:', error);
      return { success: false, error: '创建支付失败' };
    }
  }

  async handleCallback(data: any): Promise<PaymentCallbackData> {
    await this.ensureInitialized();
    
    if (!this.wechat) {
      throw new Error('微信支付未配置');
    }

    try {
      const result = await this.wechat.webhook.verify(data);

      if (!result.success) {
        throw new Error('签名验证失败');
      }

      const paymentData = result.decryptedData;
      const outTradeNo = paymentData.out_trade_no;
      const transactionId = paymentData.transaction_id;
      const amount = paymentData.amount.total / 100;
      const paidAt = new Date(paymentData.success_time);

      await updatePaymentStatus(outTradeNo, 'paid', {
        transactionId,
        paidAt,
      });

      return {
        outTradeNo,
        transactionId,
        amount,
        status: 'paid',
        paidAt,
      };
    } catch (error) {
      console.error('处理微信支付回调失败:', error);
      throw error;
    }
  }

  async queryPayment(outTradeNo: string): Promise<PaymentCallbackData | null> {
    await this.ensureInitialized();
    
    if (!this.wechat) {
      throw new Error('微信支付未配置');
    }

    try {
      const result = await this.wechat.native.query({
        out_trade_no: outTradeNo,
      });

      const payment = await getPaymentByOutTradeNo(outTradeNo);
      if (!payment) {
        return null;
      }

      const transactionState = result.trade_state;
      let status: PaymentCallbackData['status'] = 'pending';

      if (transactionState === 'SUCCESS') {
        status = 'paid';
      } else if (transactionState === 'CLOSED' || transactionState === 'PAYERROR') {
        status = 'failed';
      }

      return {
        outTradeNo,
        transactionId: result.transaction_id,
        amount: result.amount.total / 100,
        status,
        paidAt: result.success_time ? new Date(result.success_time) : undefined,
      };
    } catch (error) {
      console.error('查询微信支付失败:', error);
      return null;
    }
  }

  async closePayment(outTradeNo: string): Promise<boolean> {
    await this.ensureInitialized();
    
    if (!this.wechat) {
      throw new Error('微信支付未配置');
    }

    try {
      await this.wechat.native.close(outTradeNo);

      await updatePaymentStatus(outTradeNo, 'cancelled', {
        cancelledAt: new Date(),
      });

      return true;
    } catch (error) {
      console.error('关闭微信支付失败:', error);
      return false;
    }
  }
}

export const wechatPaymentService = new WechatPaymentService();
