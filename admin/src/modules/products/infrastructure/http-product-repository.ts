import {
  Product,
  type CreateProductInput,
  type ProductRepository,
  type UpdateProductInput,
} from '../domain/product'
import { apiFetch } from '@/shared/http'

type ProductApiItem = {
  id: string
  merchant_id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
}

function mapProduct(item: ProductApiItem): Product {
  return new Product(
    item.id,
    item.merchant_id,
    item.name,
    item.description,
    item.price_cents,
    item.currency,
    item.stock,
  )
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) {
      return body.error
    }
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

function toBody(input: CreateProductInput) {
  return {
    merchant_id: input.merchantId,
    name: input.name,
    description: input.description,
    price_cents: input.priceCents,
    currency: input.currency,
    stock: input.stock,
  }
}

export class HttpProductRepository implements ProductRepository {
  async list(): Promise<Product[]> {
    const res = await apiFetch('/api/v1/products?limit=200')
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem[] }
    return body.data.map(mapProduct)
  }

  async create(input: CreateProductInput): Promise<Product> {
    const res = await apiFetch('/api/v1/products', {
      method: 'POST',
      body: JSON.stringify(toBody(input)),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
  }

  async update(input: UpdateProductInput): Promise<Product> {
    const res = await apiFetch(`/api/v1/products/${input.id}`, {
      method: 'PUT',
      body: JSON.stringify(toBody(input)),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
  }

  async remove(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/products/${id}`, { method: 'DELETE' })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
  }
}
