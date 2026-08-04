import type {
  CreateOrderItemInput,
  CreateOrderResult,
  Order,
  OrderRepository,
  PaymentMethod,
  PaymentMethodOption,
} from '../domain/order'

export class ListMyOrdersUseCase {
  constructor(private readonly repo: OrderRepository) {}
  execute(): Promise<Order[]> {
    return this.repo.list()
  }
}

export class GetMyOrderUseCase {
  constructor(private readonly repo: OrderRepository) {}
  execute(id: string): Promise<Order> {
    return this.repo.getById(id)
  }
}

export class CreateOrderUseCase {
  constructor(private readonly repo: OrderRepository) {}
  execute(
    note: string,
    items: CreateOrderItemInput[],
    shippingName: string,
    shippingPhone: string,
    shippingAddress: string,
    paymentMethod: PaymentMethod,
  ): Promise<CreateOrderResult> {
    return this.repo.create(note, items, shippingName, shippingPhone, shippingAddress, paymentMethod)
  }
}

export class ListPaymentMethodsUseCase {
  constructor(private readonly repo: OrderRepository) {}
  execute(): Promise<PaymentMethodOption[]> {
    return this.repo.listPaymentMethods()
  }
}

export class RepayOrderUseCase {
  constructor(private readonly repo: OrderRepository) {}
  execute(orderId: string): Promise<CreateOrderResult> {
    return this.repo.repay(orderId)
  }
}
