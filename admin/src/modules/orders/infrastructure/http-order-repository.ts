import {
  DeliveryEvent,
  Order,
  OrderEvent,
  OrderItem,
  type ListOrdersFilter,
  type OrderEventType,
  type OrderRepository,
  type OrderStatus,
  type SimulateDeliveryInput,
} from '../domain/order'
import { apiFetch } from '@/shared/http'

type OrderItemApi = {
  id: string
  product_id: string
  product_name: string
  unit_price_cents: number
  quantity: number
  line_total_cents: number
}

type OrderEventApi = {
  id: string
  event_type: OrderEventType
  event_label: string
  from_status?: string
  from_status_label?: string
  to_status?: string
  to_status_label?: string
  message: string
  actor_id: string
  actor_email: string
  actor_role: string
  actor_name: string
  created_at: string
}

type DeliveryEventApi = {
  id: string
  event_id?: string
  delivery_tracking_code: string
  status_code: string
  status_label: string
  message: string
  reason?: string
  occurred_at: string
  source: string
  created_at: string
}

type OrderApiItem = {
  id: string
  code: string
  user_id: string
  merchant_id: string
  status: OrderStatus
  status_label: string
  currency: string
  total_cents: number
  note: string
  deliveryTrackingCode?: string
  deliveryCarrier?: string
  items: OrderItemApi[]
  history?: OrderEventApi[]
  delivery_events?: DeliveryEventApi[]
  created_at: string
  updated_at: string
}

function mapEvent(item: OrderEventApi): OrderEvent {
  return new OrderEvent(
    item.id,
    item.event_type,
    item.event_label,
    (item.from_status as OrderStatus) || '',
    item.from_status_label ?? '',
    (item.to_status as OrderStatus) || '',
    item.to_status_label ?? '',
    item.message,
    item.actor_id ?? '',
    item.actor_email ?? '',
    item.actor_role ?? '',
    item.actor_name ?? '',
    item.created_at,
  )
}

function mapDeliveryEvent(item: DeliveryEventApi): DeliveryEvent {
  return new DeliveryEvent(
    item.id,
    item.event_id ?? '',
    item.delivery_tracking_code ?? '',
    item.status_code,
    item.status_label,
    item.message,
    item.reason ?? '',
    item.occurred_at,
    item.source ?? '',
    item.created_at,
  )
}

function mapOrder(item: OrderApiItem): Order {
  return new Order(
    item.id,
    item.code,
    item.user_id,
    item.merchant_id,
    item.status,
    item.status_label,
    item.currency,
    item.total_cents,
    item.note ?? '',
    item.deliveryTrackingCode ?? '',
    item.deliveryCarrier ?? 'internal',
    (item.items ?? []).map(
      (line) =>
        new OrderItem(
          line.id,
          line.product_id,
          line.product_name,
          line.unit_price_cents,
          line.quantity,
          line.line_total_cents,
        ),
    ),
    (item.history ?? []).map(mapEvent),
    (item.delivery_events ?? []).map(mapDeliveryEvent),
    item.created_at,
    item.updated_at,
  )
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) {
      return body.error
    }
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

export class HttpOrderRepository implements OrderRepository {
  async list(filter: ListOrdersFilter = {}): Promise<Order[]> {
    const params = new URLSearchParams({ limit: '200' })
    if (filter.code?.trim()) {
      params.set('code', filter.code.trim().toUpperCase())
    }
    if (filter.status?.trim()) {
      params.set('status', filter.status.trim())
    }
    const res = await apiFetch(`/api/v1/orders?${params.toString()}`)
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: OrderApiItem[] }
    return body.data.map(mapOrder)
  }

  async getById(id: string): Promise<Order> {
    const res = await apiFetch(`/api/v1/orders/${id}`)
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: OrderApiItem }
    return mapOrder(body.data)
  }

  async simulateDelivery(id: string, input: SimulateDeliveryInput): Promise<Order> {
    const res = await apiFetch(`/api/v1/orders/${id}/delivery-simulate`, {
      method: 'POST',
      body: JSON.stringify({
        delivery_tracking_code: input.deliveryTrackingCode ?? '',
        delivery_carrier: input.deliveryCarrier ?? 'internal',
        status: input.status,
        message: input.message ?? '',
        reason: input.reason ?? '',
        occurred_at: input.occurredAt ?? '',
        event_id: input.eventId ?? '',
      }),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: OrderApiItem }
    return mapOrder(body.data)
  }
}
