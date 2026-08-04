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

type OrderApiItem = {
  id: string
  total_cents: number
}

export class HttpDashboardRepository implements DashboardRepository {
  async loadStats(): Promise<MerchantStats> {
    const [productsRes, ordersRes] = await Promise.all([
      apiFetch('/api/v1/merchant/products?limit=200'),
      apiFetch('/api/v1/merchant/orders?limit=200'),
    ])
    if (!productsRes.ok) {
      throw new Error(`Failed to load products (${productsRes.status})`)
    }
    if (!ordersRes.ok) {
      throw new Error(`Failed to load orders (${ordersRes.status})`)
    }
    const productsBody = (await productsRes.json()) as { data: ProductApiItem[] }
    const ordersBody = (await ordersRes.json()) as { data: OrderApiItem[] }
    const products = productsBody.data.map(
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
    const catalogValue = products.reduce((sum, p) => sum + p.priceCents, 0)
    return new MerchantStats(products.length, ordersBody.data.length, catalogValue)
  }
}
