import {
  Product,
  type CreateProductInput,
  type ProductCategory,
  type ProductRepository,
  type UpdateProductInput,
} from '../domain/product'
import { apiFetch, getAccessToken } from '@/shared/http'

type CategoryApiItem = {
  id: string
  name: string
  status: string
  status_label: string
}

type ProductApiItem = {
  id: string
  merchant_id: string
  name: string
  description: string
  price_cents: number
  currency: string
  stock: number
  image_key?: string
  image_url?: string
  categories?: CategoryApiItem[]
}

function mapCategories(items?: CategoryApiItem[]): ProductCategory[] {
  return (items ?? []).map((c) => ({
    id: c.id,
    name: c.name,
    status: c.status,
    statusLabel: c.status_label,
  }))
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
    item.image_key ?? '',
    item.image_url ?? '',
    mapCategories(item.categories),
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
    category_ids: input.categoryIds,
  }
}

function apiBase(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
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

  async getById(id: string): Promise<Product> {
    const res = await apiFetch(`/api/v1/products/${id}`)
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
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

  async removeCategory(productId: string, categoryId: string): Promise<void> {
    const res = await apiFetch(`/api/v1/products/${productId}/categories/${categoryId}`, {
      method: 'DELETE',
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
  }

  async uploadImage(id: string, file: File): Promise<Product> {
    const form = new FormData()
    form.append('file', file)
    const headers = new Headers()
    const token = getAccessToken()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
    const res = await fetch(`${apiBase()}/api/v1/products/${id}/image`, {
      method: 'POST',
      headers,
      body: form,
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
  }

  async removeImage(id: string): Promise<Product> {
    const res = await apiFetch(`/api/v1/products/${id}/image`, { method: 'DELETE' })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: ProductApiItem }
    return mapProduct(body.data)
  }
}
