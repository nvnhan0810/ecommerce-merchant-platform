import type { PaymentSettings, PaymentSettingsRepository, UpdatePaymentSettingsInput } from '../domain/payment-settings'

export class GetPaymentSettingsUseCase {
  constructor(private readonly repo: PaymentSettingsRepository) {}
  execute(): Promise<PaymentSettings> {
    return this.repo.get()
  }
}

export class UpdatePaymentSettingsUseCase {
  constructor(private readonly repo: PaymentSettingsRepository) {}
  execute(input: UpdatePaymentSettingsInput): Promise<PaymentSettings> {
    return this.repo.update(input)
  }
}
