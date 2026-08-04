export type PaymentCallbackChannel = 'ipn' | 'return'

export class PaymentCallbackEvent {
  constructor(
    readonly id: string,
    readonly provider: string,
    readonly providerLabel: string,
    readonly channel: PaymentCallbackChannel | string,
    readonly channelLabel: string,
    readonly httpMethod: string,
    readonly paymentId: string,
    readonly paymentMethod: string,
    readonly paymentMethodLabel: string,
    readonly merchTxnRef: string,
    readonly responseCode: string,
    readonly message: string,
    readonly paid: boolean,
    readonly success: boolean,
    readonly errorMessage: string,
    readonly rawPayload: Record<string, unknown>,
    readonly createdAt: string,
  ) {}
}

export const PAYMENT_PROVIDER_OPTIONS = [
  { value: 'onepay', label: 'OnePay' },
] as const

export const PAYMENT_CALLBACK_CHANNEL_OPTIONS = [
  { value: 'ipn', label: 'IPN' },
  { value: 'return', label: 'Return' },
] as const

export type ListPaymentCallbacksFilter = {
  provider?: string
  channel?: string
  merchTxnRef?: string
}

export interface PaymentCallbackRepository {
  list(filter?: ListPaymentCallbacksFilter): Promise<PaymentCallbackEvent[]>
  getById(id: string): Promise<PaymentCallbackEvent>
}
