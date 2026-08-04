import type { ListOrdersFilter, Order, OrderRepository, SimulateDeliveryInput } from '../domain/order'

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

export class SimulateDeliveryUseCase {
  constructor(private readonly repo: OrderRepository) {}

  execute(id: string, input: SimulateDeliveryInput): Promise<Order> {
    return this.repo.simulateDelivery(id, input)
  }
}
