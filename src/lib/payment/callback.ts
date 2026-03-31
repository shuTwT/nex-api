import prisma from '@/lib/prisma';
import type { PaymentStatus } from './types';

export interface PaymentCallbackData {
  outTradeNo: string;
  transactionId?: string;
  amount: number;
  status: PaymentStatus;
  paidAt?: Date;
  metadata?: Record<string, any>;
}

export interface BusinessCallbackPayload {
  paymentId: string;
  outTradeNo: string;
  transactionId?: string;
  amount: number;
  status: PaymentStatus;
  paidAt?: Date;
  userId: string;
  planId?: string;
  metadata?: Record<string, any>;
}

export async function updatePaymentRecord(
  outTradeNo: string,
  status: PaymentStatus,
  data?: {
    transactionId?: string;
    paidAt?: Date;
    cancelledAt?: Date;
    metadata?: Record<string, any>;
  }
): Promise<{ payment: any; notifyUrl: string | null } | null> {
  const payment = await prisma.payment.update({
    where: { outTradeNo },
    data: {
      status,
      transactionId: data?.transactionId,
      paidAt: data?.paidAt,
      cancelledAt: data?.cancelledAt,
      metadata: data?.metadata ? JSON.stringify(data.metadata) : undefined,
    },
    include: {
      user: true,
      plan: true,
    },
  });

  return {
    payment,
    notifyUrl: payment.notifyUrl,
  };
}

export async function callBusinessCallback(
  notifyUrl: string,
  payload: BusinessCallbackPayload
): Promise<{ success: boolean; error?: string }> {
  try {
    console.log(`调用业务回调: ${notifyUrl}`);
    console.log('回调数据:', payload);

    const response = await fetch(notifyUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`业务回调失败: ${response.status} ${errorText}`);
      return { success: false, error: `HTTP ${response.status}: ${errorText}` };
    }

    const result = await response.json();
    console.log('业务回调成功:', result);
    
    return { success: true };
  } catch (error) {
    console.error('调用业务回调异常:', error);
    return { 
      success: false, 
      error: error instanceof Error ? error.message : 'Unknown error' 
    };
  }
}

export async function processPaymentCallback(
  outTradeNo: string,
  status: PaymentStatus,
  data?: {
    transactionId?: string;
    paidAt?: Date;
    cancelledAt?: Date;
    metadata?: Record<string, any>;
  }
): Promise<{ success: boolean; error?: string }> {
  try {
    const result = await updatePaymentRecord(outTradeNo, status, data);
    
    if (!result) {
      return { success: false, error: '支付记录不存在' };
    }

    const { payment, notifyUrl } = result;

    if (notifyUrl && status === 'paid') {
      const payload: BusinessCallbackPayload = {
        paymentId: payment.id,
        outTradeNo: payment.outTradeNo,
        transactionId: payment.transactionId || undefined,
        amount: payment.amount,
        status: payment.status as PaymentStatus,
        paidAt: payment.paidAt || undefined,
        userId: payment.userId,
        planId: payment.planId || undefined,
        metadata: payment.metadata ? JSON.parse(payment.metadata) : undefined,
      };

      const callbackResult = await callBusinessCallback(notifyUrl, payload);
      
      if (!callbackResult.success) {
        console.error('业务回调失败:', callbackResult.error);
      }
    }

    return { success: true };
  } catch (error) {
    console.error('处理支付回调失败:', error);
    return { 
      success: false, 
      error: error instanceof Error ? error.message : 'Unknown error' 
    };
  }
}
