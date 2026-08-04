import type {
  OnePayDemoCredentials,
  OnePayGatewaySettings,
  PaymentSettings,
  PaymentSettingsRepository,
  UpdatePaymentSettingsInput,
} from '../domain/payment-settings'
import { apiFetch } from '@/shared/http'

type GatewayApi = {
  enabled: boolean
  merchant_id: string
  access_code: string
  hash_secret: string
  payment_url: string
  ready: boolean
}

type DemoApi = {
  merchant_id: string
  access_code: string
  hash_secret: string
  payment_url: string
  note?: string
}

type PaymentSettingsApi = {
  onepay_return_url: string
  onepay_ipn_url: string
  onepay_domestic: GatewayApi
  onepay_international: GatewayApi
  demo_domestic: DemoApi
  demo_international: DemoApi
  updated_at: string
}

function mapGateway(item: GatewayApi): OnePayGatewaySettings {
  return {
    enabled: item.enabled,
    merchantId: item.merchant_id ?? '',
    accessCode: item.access_code ?? '',
    hashSecret: item.hash_secret ?? '',
    paymentUrl: item.payment_url ?? '',
    ready: item.ready,
  }
}

function mapDemo(item: DemoApi): OnePayDemoCredentials {
  return {
    merchantId: item.merchant_id,
    accessCode: item.access_code,
    hashSecret: item.hash_secret,
    paymentUrl: item.payment_url,
    note: item.note,
  }
}

function mapSettings(item: PaymentSettingsApi): PaymentSettings {
  return {
    onepayReturnUrl: item.onepay_return_url ?? '',
    onepayIpnUrl: item.onepay_ipn_url ?? '',
    onepayDomestic: mapGateway(item.onepay_domestic),
    onepayInternational: mapGateway(item.onepay_international),
    demoDomestic: mapDemo(item.demo_domestic),
    demoInternational: mapDemo(item.demo_international),
    updatedAt: item.updated_at,
  }
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

export class HttpPaymentSettingsRepository implements PaymentSettingsRepository {
  async get(): Promise<PaymentSettings> {
    const res = await apiFetch('/api/v1/payment-settings')
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: PaymentSettingsApi }
    return mapSettings(body.data)
  }

  async update(input: UpdatePaymentSettingsInput): Promise<PaymentSettings> {
    const res = await apiFetch('/api/v1/payment-settings', {
      method: 'PUT',
      body: JSON.stringify({
        onepay_return_url: input.onepayReturnUrl,
        onepay_ipn_url: input.onepayIpnUrl,
        onepay_domestic: {
          enabled: input.onepayDomestic.enabled,
          merchant_id: input.onepayDomestic.merchantId,
          access_code: input.onepayDomestic.accessCode,
          hash_secret: input.onepayDomestic.hashSecret,
          payment_url: input.onepayDomestic.paymentUrl,
        },
        onepay_international: {
          enabled: input.onepayInternational.enabled,
          merchant_id: input.onepayInternational.merchantId,
          access_code: input.onepayInternational.accessCode,
          hash_secret: input.onepayInternational.hashSecret,
          payment_url: input.onepayInternational.paymentUrl,
        },
      }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: PaymentSettingsApi }
    return mapSettings(body.data)
  }
}
