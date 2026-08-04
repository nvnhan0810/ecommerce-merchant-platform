export class Product {
  constructor(
    readonly id: string,
    readonly name: string,
    readonly description: string,
    readonly priceCents: number,
    readonly currency: string,
    readonly stock: number,
    readonly imageKey: string = '',
    readonly imageUrl: string = '',
    readonly hasOrders: boolean = false,
    readonly canDelete: boolean = true,
    readonly categories: ProductCategory[] = [],
  ) {}
}

export type ProductCategory = {
  id: string
  name: string
  status: string
  statusLabel: string
}

export type CreateProductInput = {
  name: string
  description: string
  priceCents: number
  currency: string
  stock: number
  categoryIds: string[]
}

export type UpdateProductInput = CreateProductInput & {
  id: string
}

export interface ProductRepository {
  list(): Promise<Product[]>
  getById(id: string): Promise<Product>
  create(input: CreateProductInput): Promise<Product>
  update(input: UpdateProductInput): Promise<Product>
  remove(id: string): Promise<void>
  removeCategory(productId: string, categoryId: string): Promise<void>
  uploadImage(id: string, file: File): Promise<Product>
  removeImage(id: string): Promise<Product>
}
