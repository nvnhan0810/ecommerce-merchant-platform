import {
  Category,
  type CategoryRepository,
  type CategoryStatus,
  type CreateCategoryInput,
  type UpdateCategoryInput,
} from '../domain/category'
import { apiFetch } from '@/shared/http'

type CategoryApiItem = {
  id: string
  name: string
  status: CategoryStatus
  status_label: string
  created_by_merchant_id?: string
  created_at?: string
}

function mapCategory(item: CategoryApiItem): Category {
  return new Category(
    item.id,
    item.name,
    item.status,
    item.status_label,
    item.created_by_merchant_id ?? '',
    item.created_at ?? '',
  )
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

export class HttpCategoryRepository implements CategoryRepository {
  async list(status?: string): Promise<Category[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : ''
    const res = await apiFetch(`/api/v1/categories${qs}`)
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem[] }
    return body.data.map(mapCategory)
  }

  async getById(id: string): Promise<Category> {
    const res = await apiFetch(`/api/v1/categories/${id}`)
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem }
    return mapCategory(body.data)
  }

  async create(input: CreateCategoryInput): Promise<Category> {
    const res = await apiFetch('/api/v1/categories', {
      method: 'POST',
      body: JSON.stringify({ name: input.name }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem }
    return mapCategory(body.data)
  }

  async update(input: UpdateCategoryInput): Promise<Category> {
    const res = await apiFetch(`/api/v1/categories/${input.id}`, {
      method: 'PUT',
      body: JSON.stringify({ name: input.name }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem }
    return mapCategory(body.data)
  }

  async updateStatus(id: string, status: CategoryStatus): Promise<Category> {
    const res = await apiFetch(`/api/v1/categories/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem }
    return mapCategory(body.data)
  }

  async remove(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/categories/${id}`, { method: 'DELETE' })
    if (!res.ok) throw new Error(await readError(res))
  }
}
