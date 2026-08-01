import { Money, Product, ProductId, type MerchantProductRepository } from '../domain/product'

type ProductApiItem = {
  id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpMerchantProductRepository implements MerchantProductRepository {
  async list(): Promise<Product[]> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/products?limit=100`)
    if (!res.ok) {
      throw new Error(`Failed to load products (${res.status})`)
    }
    const body = (await res.json()) as { data: ProductApiItem[] }
    return body.data.map(
      (item) =>
        new Product(
          new ProductId(item.id),
          item.name,
          item.description,
          new Money(item.price_cents, item.currency),
          item.stock,
        ),
    )
  }
}
