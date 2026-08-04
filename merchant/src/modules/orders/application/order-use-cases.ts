import type { ListOrdersFilter, Order, OrderRepository, OrderStatus } from '../domain/order'

export class ListOrdersUseCase {
  constructor(private readonly repo: OrderRepository) {}

  execute(filter?: ListOrdersFilter): Promise<Order[]> {
    return this.repo.list(filter)
  }
}

export class GetOrderUseCase {
  constructor(private readonly repo: OrderRepository) {}

  execute(id: string): Promise<Order> {
    return this.repo.getById(id)
  }
}

export class UpdateOrderStatusUseCase {
  constructor(private readonly repo: OrderRepository) {}

  execute(id: string, status: OrderStatus, reason?: string): Promise<Order> {
    return this.repo.updateStatus(id, status, reason)
  }
}
