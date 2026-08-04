export type OnePayGatewaySettings = {
  enabled: boolean
  merchantId: string
  accessCode: string
  hashSecret: string
  paymentUrl: string
  ready: boolean
}

export type OnePayDemoCredentials = {
  merchantId: string
  accessCode: string
  hashSecret: string
  paymentUrl: string
  note?: string
}

export type PaymentSettings = {
  onepayReturnUrl: string
  onepayIpnUrl: string
  onepayDomestic: OnePayGatewaySettings
  onepayInternational: OnePayGatewaySettings
  demoDomestic: OnePayDemoCredentials
  demoInternational: OnePayDemoCredentials
  updatedAt: string
}

export type OnePayGatewayInput = {
  enabled: boolean
  merchantId: string
  accessCode: string
  hashSecret: string
  paymentUrl: string
}

export type UpdatePaymentSettingsInput = {
  onepayReturnUrl: string
  onepayIpnUrl: string
  onepayDomestic: OnePayGatewayInput
  onepayInternational: OnePayGatewayInput
}

export interface PaymentSettingsRepository {
  get(): Promise<PaymentSettings>
  update(input: UpdatePaymentSettingsInput): Promise<PaymentSettings>
}
