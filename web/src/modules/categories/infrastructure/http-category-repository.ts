import { Category, type CategoryRepository } from '../domain/category'

type CategoryApiItem = {
  id: string
  name: string
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpCategoryRepository implements CategoryRepository {
  async listApproved(): Promise<Category[]> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/categories`)
    if (!res.ok) {
      throw new Error(`Failed to load categories (${res.status})`)
    }
    const body = (await res.json()) as { data: CategoryApiItem[] }
    return body.data.map((item) => new Category(item.id, item.name))
  }
}
