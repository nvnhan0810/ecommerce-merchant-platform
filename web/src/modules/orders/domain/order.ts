export type OrderStatus =
  | 'new'
  | 'paid'
  | 'confirmed'
  | 'shipping'
  | 'succeeded'
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

export class Order {
  constructor(
    readonly id: string,
    readonly code: string,
    readonly status: OrderStatus,
    readonly statusLabel: string,
    readonly currency: string,
    readonly totalCents: number,
    readonly note: string,
    readonly items: OrderItem[],
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
  create(note: string, items: CreateOrderItemInput[]): Promise<Order[]>
}
