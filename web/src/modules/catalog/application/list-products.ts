import type { Product, ProductRepository } from '../domain/product'

export class ListProductsUseCase {
  constructor(private readonly repo: ProductRepository) {}

  async execute(limit = 20): Promise<Product[]> {
    return this.repo.list(limit)
  }
}
