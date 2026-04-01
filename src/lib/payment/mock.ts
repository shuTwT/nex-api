import { getPaymentConfig, isMockPaymentEnabled } from './config';
import type { PaymentService, CreatePaymentParams, CreatePaymentResult, PaymentCallbackData } from './types';
import { createPaymentRecord, updatePaymentStatus, getPaymentByOutTradeNo } from './utils';

class MockPaymentService implements PaymentService {
  async createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult> {
    if (!(await isMockPaymentEnabled())) {
      return { success: false, error: '模拟支付未启用' };
    }

    try {
      const config = await getPaymentConfig();
      const outTradeNo = `MOCK${Date.now()}${Math.random().toString(36).substring(2, 8)}`;
      
      const expiredAt = new Date();
      expiredAt.setHours(expiredAt.getHours() + 2);

      const mockPayUrl = `${process.env.NEXT_PUBLIC_APP_URL}/payment/mock?outTradeNo=${outTradeNo}`;

      const payment = await createPaymentRecord({
        userId: params.userId,
        outTradeNo,
        method: 'mock',
        amount: params.amount,
        payUrl: mockPayUrl,
        notifyUrl: params.notifyUrl,
        expiredAt,
        metadata: params.metadata,
      });

      if (config.mock.autoSuccess) {
        setTimeout(async () => {
          await this.handleCallback({ outTradeNo, success: true });
        }, config.mock.delay);
      }

      return {
        success: true,
        paymentId: payment.id,
        outTradeNo,
        payUrl: mockPayUrl,
      };
    } catch (error) {
      console.error('创建模拟支付失败:', error);
      return { success: false, error: '创建支付失败' };
    }
  }

  async handleCallback(data: any): Promise<PaymentCallbackData> {
    try {
      const { outTradeNo, success } = data;
      
      const payment = await getPaymentByOutTradeNo(outTradeNo);
      if (!payment) {
        throw new Error('支付记录不存在');
      }

      const transactionId = `MOCK${Date.now()}`;
      const paidAt = new Date();

      if (success) {
        await updatePaymentStatus(outTradeNo, 'paid', {
          transactionId,
          paidAt,
        });

        return {
          outTradeNo,
          transactionId,
          amount: payment.amount,
          status: 'paid',
          paidAt,
        };
      } else {
        await updatePaymentStatus(outTradeNo, 'failed');

        return {
          outTradeNo,
          amount: payment.amount,
          status: 'failed',
        };
      }
    } catch (error) {
      console.error('处理模拟支付回调失败:', error);
      throw error;
    }
  }

  async queryPayment(outTradeNo: string): Promise<PaymentCallbackData | null> {
    try {
      const payment = await getPaymentByOutTradeNo(outTradeNo);
      if (!payment) {
        return null;
      }

      return {
        outTradeNo,
        transactionId: payment.transactionId || undefined,
        amount: payment.amount,
        status: payment.status as PaymentCallbackData['status'],
        paidAt: payment.paidAt || undefined,
      };
    } catch (error) {
      console.error('查询模拟支付失败:', error);
      return null;
    }
  }

  async closePayment(outTradeNo: string): Promise<boolean> {
    try {
      await updatePaymentStatus(outTradeNo, 'cancelled', {
        cancelledAt: new Date(),
      });

      return true;
    } catch (error) {
      console.error('关闭模拟支付失败:', error);
      return false;
    }
  }
}

export const mockPaymentService = new MockPaymentService();
