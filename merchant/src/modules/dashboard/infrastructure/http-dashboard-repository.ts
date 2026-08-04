import { Product } from '../../products/domain/product'
import { MerchantStats, type DashboardRepository } from '../domain/stats'
import { apiFetch } from '@/shared/http'

type ProductApiItem = {
  id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
  image_key?: string
  image_url?: string
  has_orders?: boolean
  can_delete?: boolean
}

export class HttpDashboardRepository implements DashboardRepository {
  async loadStats(): Promise<MerchantStats> {
    const res = await apiFetch('/api/v1/merchant/products?limit=200')
    if (!res.ok) {
      throw new Error(`Failed to load dashboard (${res.status})`)
    }
    const body = (await res.json()) as { data: ProductApiItem[] }
    const products = body.data.map(
      (item) =>
        new Product(
          item.id,
          item.name,
          item.description,
          item.price_cents,
          item.currency,
          item.stock,
          item.image_key ?? '',
          item.image_url ?? '',
          Boolean(item.has_orders),
          item.can_delete ?? !item.has_orders,
        ),
    )
    const revenueCents = products.reduce((sum, p) => sum + p.priceCents, 0)
    return new MerchantStats(products.length, 0, revenueCents)
  }
}
