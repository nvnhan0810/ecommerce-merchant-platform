import { Money, Product, ProductId } from '../../products/domain/product'
import { MerchantStats, type DashboardRepository } from '../domain/stats'

type ProductApiItem = {
  id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
  merchant_id: string
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpDashboardRepository implements DashboardRepository {
  async loadStats(): Promise<MerchantStats> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/products?limit=100`)
    if (!res.ok) {
      throw new Error(`Failed to load dashboard (${res.status})`)
    }
    const body = (await res.json()) as { data: ProductApiItem[] }
    const products = body.data.map(
      (item) =>
        new Product(
          new ProductId(item.id),
          item.name,
          item.description,
          new Money(item.price_cents, item.currency),
          item.stock,
        ),
    )
    const revenueCents = products.reduce((sum, p) => sum + p.price.amountCents, 0)
    return new MerchantStats(products.length, 0, revenueCents)
  }
}
