import type { MerchantProductRepository, Product } from '../domain/product'

export class ListMerchantProductsUseCase {
  constructor(private readonly repo: MerchantProductRepository) {}

  execute(): Promise<Product[]> {
    return this.repo.list()
  }
}
