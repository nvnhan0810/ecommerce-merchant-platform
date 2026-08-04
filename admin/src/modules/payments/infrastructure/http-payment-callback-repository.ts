import {
  PaymentCallbackEvent,
  type ListPaymentCallbacksFilter,
  type PaymentCallbackRepository,
} from '../domain/payment-callback'
import { apiFetch } from '@/shared/http'

type PaymentCallbackApi = {
  id: string
  provider: string
  provider_label: string
  channel: string
  channel_label: string
  http_method: string
  payment_id?: string
  payment_method?: string
  payment_method_label?: string
  merch_txn_ref: string
  response_code: string
  message: string
  paid: boolean
  success: boolean
  error_message?: string
  raw_payload?: Record<string, unknown> | string
  created_at: string
}

function mapRawPayload(raw: PaymentCallbackApi['raw_payload']): Record<string, unknown> {
  if (!raw) return {}
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw) as Record<string, unknown>
    } catch {
      return { raw }
    }
  }
  return raw
}

function mapEvent(item: PaymentCallbackApi): PaymentCallbackEvent {
  return new PaymentCallbackEvent(
    item.id,
    item.provider,
    item.provider_label || item.provider,
    item.channel,
    item.channel_label || item.channel,
    item.http_method || '',
    item.payment_id ?? '',
    item.payment_method ?? '',
    item.payment_method_label ?? '',
    item.merch_txn_ref ?? '',
    item.response_code ?? '',
    item.message ?? '',
    Boolean(item.paid),
    Boolean(item.success),
    item.error_message ?? '',
    mapRawPayload(item.raw_payload),
    item.created_at,
  )
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

export class HttpPaymentCallbackRepository implements PaymentCallbackRepository {
  async list(filter: ListPaymentCallbacksFilter = {}): Promise<PaymentCallbackEvent[]> {
    const params = new URLSearchParams()
    params.set('limit', '200')
    if (filter.provider) params.set('provider', filter.provider)
    if (filter.channel) params.set('channel', filter.channel)
    if (filter.merchTxnRef) params.set('merch_txn_ref', filter.merchTxnRef)
    const res = await apiFetch(`/api/v1/payment-callbacks?${params.toString()}`)
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: PaymentCallbackApi[] }
    return (body.data ?? []).map(mapEvent)
  }

  async getById(id: string): Promise<PaymentCallbackEvent> {
    const res = await apiFetch(`/api/v1/payment-callbacks/${id}`)
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: PaymentCallbackApi }
    return mapEvent(body.data)
  }
}
