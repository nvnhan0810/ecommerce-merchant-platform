import {
  UserAccount,
  type CreateUserInput,
  type UpdateUserInput,
  type UserRepository,
} from '../domain/user'
import { apiFetch } from '@/shared/http'

type UserApiItem = {
  id: string
  email: string
  display_name: string
  role: string
}

function mapUser(item: UserApiItem): UserAccount {
  return new UserAccount(item.id, item.email, item.display_name, item.role)
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

export class HttpUserRepository implements UserRepository {
  async list(): Promise<UserAccount[]> {
    const res = await apiFetch('/api/v1/users')
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: UserApiItem[] }
    return body.data.map(mapUser)
  }

  async getById(id: string): Promise<UserAccount> {
    const res = await apiFetch(`/api/v1/users/${id}`)
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: UserApiItem }
    return mapUser(body.data)
  }

  async create(input: CreateUserInput): Promise<UserAccount> {
    const res = await apiFetch('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify({
        email: input.email,
        display_name: input.displayName,
        password: input.password,
      }),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: UserApiItem }
    return mapUser(body.data)
  }

  async update(input: UpdateUserInput): Promise<UserAccount> {
    const res = await apiFetch(`/api/v1/users/${input.id}`, {
      method: 'PUT',
      body: JSON.stringify({
        email: input.email,
        display_name: input.displayName,
        password: input.password ?? '',
      }),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: UserApiItem }
    return mapUser(body.data)
  }

  async remove(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
  }
}
