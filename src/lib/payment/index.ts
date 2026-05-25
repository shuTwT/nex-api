import type { PaymentMethod, PaymentService, CreatePaymentParams, CreatePaymentResult, PaymentCallbackData } from './types';
import { wechatPaymentService } from './wechat';
import { alipayPaymentService } from './alipay';
import { mockPaymentService } from './mock';

export class PaymentServiceFactory {
  private static services: Record<PaymentMethod, PaymentService> = {
    wechat: wechatPaymentService,
    alipay: alipayPaymentService,
    mock: mockPaymentService,
  };

  static getService(method: PaymentMethod): PaymentService {
    const service = this.services[method];
    if (!service) {
      throw new Error(`不支持的支付方式: ${method}`);
    }
    return service;
  }

  static async createPayment(params: CreatePaymentParams): Promise<CreatePaymentResult> {
    const service = this.getService(params.method);
    return await service.createPayment(params);
  }

  static async handleCallback(method: PaymentMethod, data: unknown): Promise<PaymentCallbackData> {
    const service = this.getService(method);
    return await service.handleCallback(data);
  }

  static async queryPayment(method: PaymentMethod, outTradeNo: string): Promise<PaymentCallbackData | null> {
    const service = this.getService(method);
    return await service.queryPayment(outTradeNo);
  }

  static async closePayment(method: PaymentMethod, outTradeNo: string): Promise<boolean> {
    const service = this.getService(method);
    return await service.closePayment(outTradeNo);
  }
}

export { wechatPaymentService, alipayPaymentService, mockPaymentService };
export * from './types';
export * from './config';
export * from './utils';
