export type OrderStatus =
  | 'new'
  | 'paid'
  | 'confirmed'
  | 'shipping'
  | 'succeeded'
  | 'returning'
  | 'returned'
  | 'failed'
  | 'cancelled'

export type OrderEventType = 'created' | 'status_changed' | 'cancelled'

export type DeliveryStatusCode =
  | 'accepted'
  | 'delivering'
  | 'in_transit'
  | 'delivery_fail'
  | 'delivered'
  | 'returning'
  | 'returned'
  | 'return_fail'
  | 'lost'
  | 'damage'

export const ORDER_STATUS_OPTIONS: { value: OrderStatus; label: string }[] = [
  { value: 'new', label: 'Mới' },
  { value: 'paid', label: 'Đã thanh toán' },
  { value: 'confirmed', label: 'Đã xác nhận' },
  { value: 'shipping', label: 'Đang vận chuyển' },
  { value: 'succeeded', label: 'Thành công' },
  { value: 'returning', label: 'Đang hoàn hàng' },
  { value: 'returned', label: 'Đã hoàn hàng' },
  { value: 'failed', label: 'Thất bại' },
  { value: 'cancelled', label: 'Huỷ' },
]

export const DELIVERY_STATUS_OPTIONS: { value: DeliveryStatusCode; label: string }[] = [
  { value: 'accepted', label: 'Đã tiếp nhận' },
  { value: 'delivering', label: 'Đang giao' },
  { value: 'in_transit', label: 'Đang trung chuyển' },
  { value: 'delivery_fail', label: 'Giao thất bại' },
  { value: 'delivered', label: 'Đã giao' },
  { value: 'returning', label: 'Đang hoàn hàng' },
  { value: 'returned', label: 'Đã hoàn hàng' },
  { value: 'return_fail', label: 'Hoàn thất bại' },
  { value: 'lost', label: 'Thất lạc' },
  { value: 'damage', label: 'Hư hỏng' },
]

export class OrderItem {
  constructor(
    readonly id: string,
    readonly productId: string,
    readonly productName: string,
    readonly unitPriceCents: number,
    readonly quantity: number,
    readonly lineTotalCents: number,
  ) {}
}

export class OrderEvent {
  constructor(
    readonly id: string,
    readonly eventType: OrderEventType,
    readonly eventLabel: string,
    readonly fromStatus: OrderStatus | '',
    readonly fromStatusLabel: string,
    readonly toStatus: OrderStatus | '',
    readonly toStatusLabel: string,
    readonly message: string,
    readonly actorId: string,
    readonly actorEmail: string,
    readonly actorRole: string,
    readonly actorName: string,
    readonly createdAt: string,
  ) {}
}

export class DeliveryEvent {
  constructor(
    readonly id: string,
    readonly eventId: string,
    readonly deliveryTrackingCode: string,
    readonly statusCode: string,
    readonly statusLabel: string,
    readonly message: string,
    readonly reason: string,
    readonly occurredAt: string,
    readonly source: string,
    readonly createdAt: string,
  ) {}
}

export class Order {
  constructor(
    readonly id: string,
    readonly code: string,
    readonly userId: string,
    readonly merchantId: string,
    readonly status: OrderStatus,
    readonly statusLabel: string,
    readonly currency: string,
    readonly totalCents: number,
    readonly note: string,
    readonly deliveryTrackingCode: string,
    readonly deliveryCarrier: string,
    readonly shippingName: string,
    readonly shippingPhone: string,
    readonly shippingAddress: string,
    readonly items: OrderItem[],
    readonly history: OrderEvent[],
    readonly deliveryEvents: DeliveryEvent[],
    readonly createdAt: string,
    readonly updatedAt: string,
  ) {}
}

export type ListOrdersFilter = {
  code?: string
  status?: string
}

export type SimulateDeliveryInput = {
  deliveryTrackingCode?: string
  deliveryCarrier?: string
  status: DeliveryStatusCode
  message?: string
  reason?: string
  occurredAt?: string
  eventId?: string
}

export interface OrderRepository {
  list(filter?: ListOrdersFilter): Promise<Order[]>
  getById(id: string): Promise<Order>
  simulateDelivery(id: string, input: SimulateDeliveryInput): Promise<Order>
}
