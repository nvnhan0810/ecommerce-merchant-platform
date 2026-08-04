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

export class GetProductUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(id: string): Promise<Product> {
    return this.repo.getById(id)
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

export class RemoveProductCategoryUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(productId: string, categoryId: string): Promise<void> {
    return this.repo.removeCategory(productId, categoryId)
  }
}

export class UploadProductImageUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(id: string, file: File): Promise<Product> {
    return this.repo.uploadImage(id, file)
  }
}

export class DeleteProductImageUseCase {
  constructor(private readonly repo: ProductRepository) {}

  execute(id: string): Promise<Product> {
    return this.repo.removeImage(id)
  }
}
