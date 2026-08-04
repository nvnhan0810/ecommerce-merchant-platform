import { Money, Product, ProductId, type ProductRepository } from '../domain/product'

type ProductApiItem = {
  id: string
  merchant_id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
  image_url?: string
  merchant_display_name?: string
  merchant_avatar_url?: string
  merchant_province_name?: string
}

function mapProduct(item: ProductApiItem): Product {
  return new Product(
    new ProductId(item.id),
    item.name,
    item.description,
    new Money(item.price_cents, item.currency),
    item.stock,
    item.merchant_id,
    item.image_url ?? '',
    item.merchant_display_name ?? '',
    item.merchant_avatar_url ?? '',
    item.merchant_province_name ?? '',
  )
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpProductRepository implements ProductRepository {
  async list(limit = 40, merchantId?: string): Promise<Product[]> {
    const params = new URLSearchParams({ limit: String(limit) })
    if (merchantId?.trim()) {
      params.set('merchant_id', merchantId.trim())
    }
    const res = await fetch(`${apiBaseUrl()}/api/v1/products?${params.toString()}`)
    if (!res.ok) {
      throw new Error(`Failed to load products (${res.status})`)
    }
    const body = (await res.json()) as { data: ProductApiItem[] }
    return body.data.map(mapProduct)
  }

  async getById(id: string): Promise<Product> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/products/${id}`)
    if (!res.ok) {
      throw new Error(`Failed to load product (${res.status})`)
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
  }
}
