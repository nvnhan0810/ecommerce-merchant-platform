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

export class DeliveryEvent {
  constructor(
    readonly id: string,
    readonly deliveryTrackingCode: string,
    readonly statusCode: string,
    readonly statusLabel: string,
    readonly message: string,
    readonly reason: string,
    readonly occurredAt: string,
    readonly source: string,
  ) {}
}

export class Order {
  constructor(
    readonly id: string,
    readonly code: string,
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
    readonly deliveryEvents: DeliveryEvent[],
    readonly createdAt: string,
    readonly updatedAt: string,
  ) {}
}

export type CreateOrderItemInput = {
  productId: string
  quantity: number
}

export interface OrderRepository {
  list(): Promise<Order[]>
  getById(id: string): Promise<Order>
  create(note: string, items: CreateOrderItemInput[], shippingName: string, shippingPhone: string, shippingAddress: string): Promise<Order[]>
}
