export class Product {
  constructor(
    readonly id: string,
    readonly merchantId: string,
    readonly name: string,
    readonly description: string,
    readonly priceCents: number,
    readonly currency: string,
    readonly stock: number,
  ) {}
}

export type CreateProductInput = {
  merchantId: string
  name: string
  description: string
  priceCents: number
  currency: string
  stock: number
}

export type UpdateProductInput = CreateProductInput & {
  id: string
}

export interface ProductRepository {
  list(): Promise<Product[]>
  create(input: CreateProductInput): Promise<Product>
  update(input: UpdateProductInput): Promise<Product>
  remove(id: string): Promise<void>
}
