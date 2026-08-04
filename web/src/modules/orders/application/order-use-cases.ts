import type { CreateOrderItemInput, Order, OrderRepository } from '../domain/order'

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
  execute(note: string, items: CreateOrderItemInput[]): Promise<Order[]> {
    return this.repo.create(note, items)
  }
}
