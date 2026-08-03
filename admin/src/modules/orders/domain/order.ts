export type OrderStatus =
  | 'new'
  | 'paid'
  | 'confirmed'
  | 'shipping'
  | 'succeeded'
  | 'failed'
  | 'cancelled'

export type OrderEventType = 'created' | 'status_changed' | 'cancelled'

export const ORDER_STATUS_OPTIONS: { value: OrderStatus; label: string }[] = [
  { value: 'new', label: 'Mới' },
  { value: 'paid', label: 'Đã thanh toán' },
  { value: 'confirmed', label: 'Đã xác nhận' },
  { value: 'shipping', label: 'Đang vận chuyển' },
  { value: 'succeeded', label: 'Thành công' },
  { value: 'failed', label: 'Thất bại' },
  { value: 'cancelled', label: 'Huỷ' },
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
    readonly items: OrderItem[],
    readonly history: OrderEvent[],
    readonly createdAt: string,
    readonly updatedAt: string,
  ) {}
}

export type ListOrdersFilter = {
  code?: string
  status?: string
}

export interface OrderRepository {
  list(filter?: ListOrdersFilter): Promise<Order[]>
  getById(id: string): Promise<Order>
  updateStatus(id: string, status: OrderStatus): Promise<Order>
}
