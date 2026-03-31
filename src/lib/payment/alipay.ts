import { AlipaySdk } from 'alipay-sdk';
import { paymentConfig, isAlipayConfigured } from './config';
import type { PaymentService, CreatePaymentParams, CreatePaymentResult, PaymentCallbackData } from './types';
import { createPaymentRecord, updatePaymentStatus, getPaymentByOutTradeNo } from './utils';

class AlipayPaymentService implements PaymentService {
  private alipay: AlipaySdk | null = null;

  constructor() {
    if (isAlipayConfigured()) {
      this.alipay = new AlipaySdk({
        appId: paymentConfig.alipay.appId,
        privateKey: paymentConfig.alipay.privateKey,
        alipayPublicKey: paymentConfig.alipay.alipayPublicKey,
        gateway: paymentConfig.alipay.sandbox 
          ? 'https://openapi.alipaydev.com/gateway.do' 
          : 'https://openapi.alipay.com/gateway.do',
      });
    }
  }

  async createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult> {
    if (!this.alipay) {
      return { success: false, error: '支付宝支付未配置' };
    }

    try {
      const outTradeNo = `ALI${Date.now()}${Math.random().toString(36).substring(2, 8)}`;
      
      const result = this.alipay.pageExec(
        'alipay.trade.page.pay',
        {
          notify_url: paymentConfig.alipay.notifyUrl,
          return_url: paymentConfig.alipay.returnUrl,
          bizContent: {
            out_trade_no: outTradeNo,
            product_code: 'FAST_INSTANT_TRADE_PAY',
            total_amount: params.amount.toFixed(2),
            subject: `订阅计划-${params.planId}`,
          },
        }
      );

      const expiredAt = new Date();
      expiredAt.setHours(expiredAt.getHours() + 2);

      const payment = await createPaymentRecord({
        userId: params.userId,
        planId: params.planId,
        outTradeNo,
        method: 'alipay',
        amount: params.amount,
        payUrl: result,
        notifyUrl: params.notifyUrl,
        expiredAt,
        metadata: params.metadata,
      });

      return {
        success: true,
        paymentId: payment.id,
        outTradeNo,
        payUrl: result,
      };
    } catch (error) {
      console.error('创建支付宝支付失败:', error);
      return { success: false, error: '创建支付失败' };
    }
  }

  async handleCallback(data: any): Promise<PaymentCallbackData> {
    if (!this.alipay) {
      throw new Error('支付宝支付未配置');
    }

    try {
      const signVerified = this.alipay.checkNotifySign(data);
      
      if (!signVerified) {
        throw new Error('签名验证失败');
      }

      const outTradeNo = data.out_trade_no;
      const transactionId = data.trade_no;
      const amount = parseFloat(data.total_amount);
      const tradeStatus = data.trade_status;
      const paidAt = new Date(data.gmt_payment);

      if (tradeStatus === 'TRADE_SUCCESS' || tradeStatus === 'TRADE_FINISHED') {
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
      } else {
        await updatePaymentStatus(outTradeNo, 'failed');
        
        return {
          outTradeNo,
          transactionId,
          amount,
          status: 'failed',
        };
      }
    } catch (error) {
      console.error('处理支付宝支付回调失败:', error);
      throw error;
    }
  }

  async queryPayment(outTradeNo: string): Promise<PaymentCallbackData | null> {
    if (!this.alipay) {
      throw new Error('支付宝支付未配置');
    }

    try {
      const result = await this.alipay.exec(
        'alipay.trade.query',
        {
          bizContent: {
            out_trade_no: outTradeNo,
          },
        }
      );

      const payment = await getPaymentByOutTradeNo(outTradeNo);
      if (!payment) {
        return null;
      }

      const tradeStatus = result.tradeStatus;
      let status: PaymentCallbackData['status'] = 'pending';

      if (tradeStatus === 'TRADE_SUCCESS' || tradeStatus === 'TRADE_FINISHED') {
        status = 'paid';
      } else if (tradeStatus === 'TRADE_CLOSED') {
        status = 'failed';
      }

      return {
        outTradeNo,
        transactionId: result.tradeNo,
        amount: parseFloat(result.totalAmount),
        status,
        paidAt: result.sendPayDate ? new Date(result.sendPayDate) : undefined,
      };
    } catch (error) {
      console.error('查询支付宝支付失败:', error);
      return null;
    }
  }

  async closePayment(outTradeNo: string): Promise<boolean> {
    if (!this.alipay) {
      throw new Error('支付宝支付未配置');
    }

    try {
      await this.alipay.exec(
        'alipay.trade.close',
        {
          bizContent: {
            out_trade_no: outTradeNo,
          },
        }
      );

      await updatePaymentStatus(outTradeNo, 'cancelled', {
        cancelledAt: new Date(),
      });

      return true;
    } catch (error) {
      console.error('关闭支付宝支付失败:', error);
      return false;
    }
  }
}

export const alipayPaymentService = new AlipayPaymentService();
