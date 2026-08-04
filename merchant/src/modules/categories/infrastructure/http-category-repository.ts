import { Category, type CategoryRepository, type CategoryStatus } from '../domain/category'
import { apiFetch } from '@/shared/http'

type CategoryApiItem = {
  id: string
  name: string
  status: CategoryStatus
  status_label: string
}

function mapCategory(item: CategoryApiItem): Category {
  return new Category(item.id, item.name, item.status, item.status_label)
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
  async listAssignable(): Promise<Category[]> {
    const res = await apiFetch('/api/v1/merchant/categories')
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem[] }
    return body.data.map(mapCategory)
  }

  async create(name: string): Promise<Category> {
    const res = await apiFetch('/api/v1/merchant/categories', {
      method: 'POST',
      body: JSON.stringify({ name }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: CategoryApiItem }
    return mapCategory(body.data)
  }
}
