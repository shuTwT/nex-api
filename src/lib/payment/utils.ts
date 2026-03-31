import prisma from '@/lib/prisma';
import type { Payment } from '../../../generated/client';
import type { PaymentStatus } from './types';

export function generateOutTradeNo(): string {
  const timestamp = Date.now().toString(36).toUpperCase();
  const random = Math.random().toString(36).substring(2, 8).toUpperCase();
  return `PAY${timestamp}${random}`;
}

export async function createPaymentRecord(params: {
  userId: string;
  planId?: string;
  outTradeNo: string;
  method: string;
  amount: number;
  qrcodeUrl?: string;
  payUrl?: string;
  notifyUrl?: string;
  expiredAt?: Date;
  metadata?: Record<string, any>;
}): Promise<Payment> {
  return await prisma.payment.create({
    data: {
      userId: params.userId,
      planId: params.planId,
      outTradeNo: params.outTradeNo,
      method: params.method,
      amount: params.amount,
      qrcodeUrl: params.qrcodeUrl,
      payUrl: params.payUrl,
      notifyUrl: params.notifyUrl,
      expiredAt: params.expiredAt,
      metadata: params.metadata ? JSON.stringify(params.metadata) : null,
      status: 'pending',
    },
  });
}

export async function updatePaymentStatus(
  outTradeNo: string,
  status: PaymentStatus,
  data?: {
    transactionId?: string;
    paidAt?: Date;
    cancelledAt?: Date;
    metadata?: Record<string, any>;
  }
): Promise<Payment | null> {
  return await prisma.payment.update({
    where: { outTradeNo },
    data: {
      status,
      transactionId: data?.transactionId,
      paidAt: data?.paidAt,
      cancelledAt: data?.cancelledAt,
      metadata: data?.metadata ? JSON.stringify(data.metadata) : undefined,
    },
  });
}

export async function getPaymentByOutTradeNo(outTradeNo: string): Promise<Payment | null> {
  return await prisma.payment.findUnique({
    where: { outTradeNo },
    include: {
      user: true,
      plan: true,
    },
  });
}

export async function getPaymentById(id: string): Promise<Payment | null> {
  return await prisma.payment.findUnique({
    where: { id },
    include: {
      user: true,
      plan: true,
    },
  });
}

export async function getUserPayments(userId: string): Promise<Payment[]> {
  return await prisma.payment.findMany({
    where: { userId },
    include: {
      plan: true,
    },
    orderBy: { createdAt: 'desc' },
  });
}

export async function getPendingPayments(userId: string): Promise<Payment[]> {
  return await prisma.payment.findMany({
    where: {
      userId,
      status: 'pending',
      expiredAt: {
        gte: new Date(),
      },
    },
    include: {
      plan: true,
    },
    orderBy: { createdAt: 'desc' },
  });
}
