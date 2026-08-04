export type OrderStatus =
  | 'awaiting_payment'
  | 'new'
  | 'paid'
  | 'confirmed'
  | 'shipping'
  | 'succeeded'
  | 'returning'
  | 'returned'
  | 'failed'
  | 'cancelled'

export type PaymentMethod = 'cod' | 'onepay_domestic' | 'onepay_international' | 'onepay'
export type PaymentStatus = 'unpaid' | 'pending' | 'paid' | 'failed' | 'cancelled'

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
    readonly paymentMethod: PaymentMethod,
    readonly paymentMethodLabel: string,
    readonly paymentStatus: PaymentStatus,
    readonly paymentStatusLabel: string,
    readonly paymentId: string,
    readonly canRepay: boolean,
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

export type PaymentInfo = {
  id: string
  method: PaymentMethod
  methodLabel: string
  status: PaymentStatus
  statusLabel: string
  amountCents: number
  currency: string
  paymentUrl: string
}

export type CreateOrderResult = {
  orders: Order[]
  payment?: PaymentInfo
}

export type PaymentMethodOption = {
  method: PaymentMethod
  label: string
  enabled: boolean
  description: string
}

export interface OrderRepository {
  list(): Promise<Order[]>
  getById(id: string): Promise<Order>
  create(
    note: string,
    items: CreateOrderItemInput[],
    shippingName: string,
    shippingPhone: string,
    shippingAddress: string,
    paymentMethod: PaymentMethod,
  ): Promise<CreateOrderResult>
  repay(orderId: string): Promise<CreateOrderResult>
  listPaymentMethods(): Promise<PaymentMethodOption[]>
}
