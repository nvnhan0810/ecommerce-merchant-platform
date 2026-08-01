import { Money, Product, ProductId, type ProductRepository } from '../domain/product'

type ProductApiItem = {
  id: string
  merchant_id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
}

type ListProductsResponse = {
  data: ProductApiItem[]
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpProductRepository implements ProductRepository {
  async list(limit = 20): Promise<Product[]> {
    const url = `${apiBaseUrl()}/api/v1/products?limit=${limit}`
    const res = await fetch(url)
    if (!res.ok) {
      throw new Error(`Failed to load products (${res.status})`)
    }
    const body = (await res.json()) as ListProductsResponse
    return body.data.map(
      (item) =>
        new Product(
          new ProductId(item.id),
          item.name,
          item.description,
          new Money(item.price_cents, item.currency),
          item.stock,
          item.merchant_id,
        ),
    )
  }
}
