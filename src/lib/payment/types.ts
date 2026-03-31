export type PaymentMethod = 'wechat' | 'alipay' | 'mock';

export type PaymentStatus = 'pending' | 'paid' | 'failed' | 'cancelled' | 'expired';

export interface CreatePaymentParams {
  userId: string;
  planId: string;
  amount: number;
  method: PaymentMethod;
  notifyUrl?: string;
  metadata?: Record<string, any>;
}

export interface CreatePaymentResult {
  success: boolean;
  paymentId?: string;
  outTradeNo?: string;
  qrcodeUrl?: string;
  payUrl?: string;
  error?: string;
}

export interface PaymentCallbackData {
  outTradeNo: string;
  transactionId?: string;
  amount: number;
  status: PaymentStatus;
  paidAt?: Date;
  metadata?: Record<string, any>;
}

export interface PaymentService {
  createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult>;
  handleCallback(data: any): Promise<PaymentCallbackData>;
  queryPayment(outTradeNo: string): Promise<PaymentCallbackData | null>;
  closePayment(outTradeNo: string): Promise<boolean>;
}

export interface PaymentInfo {
  id: string;
  outTradeNo: string;
  transactionId?: string;
  method: PaymentMethod;
  amount: number;
  status: PaymentStatus;
  qrcodeUrl?: string;
  payUrl?: string;
  paidAt?: Date;
  expiredAt?: Date;
  createdAt: Date;
}
