import type { ListPaymentCallbacksFilter, PaymentCallbackRepository } from '../domain/payment-callback'
import type { PaymentCallbackEvent } from '../domain/payment-callback'

export class ListPaymentCallbacksUseCase {
  constructor(private readonly repo: PaymentCallbackRepository) {}
  execute(filter?: ListPaymentCallbacksFilter): Promise<PaymentCallbackEvent[]> {
    return this.repo.list(filter)
  }
}

export class GetPaymentCallbackUseCase {
  constructor(private readonly repo: PaymentCallbackRepository) {}
  execute(id: string): Promise<PaymentCallbackEvent> {
    return this.repo.getById(id)
  }
}
