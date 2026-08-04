import type { Product, ProductRepository } from '../domain/product'

export class ListProductsUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(limit?: number, merchantId?: string): Promise<Product[]> {
    return this.repo.list(limit, merchantId)
  }
}

export class GetProductUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(id: string): Promise<Product> {
    return this.repo.getById(id)
  }
}
