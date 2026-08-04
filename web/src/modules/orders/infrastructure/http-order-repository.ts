import {
  DeliveryEvent,
  Order,
  OrderItem,
  type CreateOrderItemInput,
  type OrderRepository,
  type OrderStatus,
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

type DeliveryEventApi = {
  id: string
  delivery_tracking_code: string
  status_code: string
  status_label: string
  message: string
  reason?: string
  occurred_at: string
  source: string
}

type OrderApiItem = {
  id: string
  code: string
  status: OrderStatus
  status_label: string
  currency: string
  total_cents: number
  note: string
  deliveryTrackingCode?: string
  deliveryCarrier?: string
  shipping_name?: string
  shipping_phone?: string
  shipping_address?: string
  items: OrderItemApi[]
  delivery_events?: DeliveryEventApi[]
  history?: unknown[]
  created_at: string
  updated_at: string
}

function mapOrder(item: OrderApiItem): Order {
  return new Order(
    item.id,
    item.code,
    item.status,
    item.status_label,
    item.currency,
    item.total_cents,
    item.note ?? '',
    item.deliveryTrackingCode ?? '',
    item.deliveryCarrier ?? 'internal',
    item.shipping_name ?? '',
    item.shipping_phone ?? '',
    item.shipping_address ?? '',
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
    (item.delivery_events ?? []).map(
      (ev) =>
        new DeliveryEvent(
          ev.id,
          ev.delivery_tracking_code ?? '',
          ev.status_code,
          ev.status_label,
          ev.message,
          ev.reason ?? '',
          ev.occurred_at,
          ev.source ?? '',
        ),
    ),
    item.created_at,
    item.updated_at,
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

export class HttpOrderRepository implements OrderRepository {
  async list(): Promise<Order[]> {
    const res = await apiFetch('/api/v1/me/orders?limit=100')
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: OrderApiItem[] }
    return body.data.map(mapOrder)
  }

  async getById(id: string): Promise<Order> {
    const res = await apiFetch(`/api/v1/me/orders/${id}`)
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: OrderApiItem }
    return mapOrder(body.data)
  }

  async create(note: string, items: CreateOrderItemInput[], shippingName: string, shippingPhone: string, shippingAddress: string): Promise<Order[]> {
    const res = await apiFetch('/api/v1/orders', {
      method: 'POST',
      body: JSON.stringify({
        note,
        shipping_name: shippingName,
        shipping_phone: shippingPhone,
        shipping_address: shippingAddress,
        items: items.map((item) => ({
          product_id: item.productId,
          quantity: item.quantity,
        })),
      }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: OrderApiItem[] }
    return body.data.map(mapOrder)
  }
}
