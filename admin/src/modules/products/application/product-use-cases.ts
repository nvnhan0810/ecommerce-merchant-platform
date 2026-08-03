import type {
  CreateProductInput,
  Product,
  ProductRepository,
  UpdateProductInput,
} from '../domain/product'

export class ListProductsUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(): Promise<Product[]> {
    return this.repo.list()
  }
}

export class CreateProductUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(input: CreateProductInput): Promise<Product> {
    return this.repo.create(input)
  }
}

export class UpdateProductUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(input: UpdateProductInput): Promise<Product> {
    return this.repo.update(input)
  }
}

export class DeleteProductUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(id: string): Promise<void> {
    return this.repo.remove(id)
  }
}
