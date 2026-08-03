import {
  MerchantAccount,
  type CreateMerchantInput,
  type MerchantRepository,
  type UpdateMerchantInput,
} from '../domain/merchant'
import { apiFetch } from '@/shared/http'

type MerchantApiItem = {
  id: string
  email: string
  display_name: string
  role: string
}

function mapMerchant(item: MerchantApiItem): MerchantAccount {
  return new MerchantAccount(item.id, item.email, item.display_name, item.role)
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

export class HttpMerchantRepository implements MerchantRepository {
  async list(): Promise<MerchantAccount[]> {
    const res = await apiFetch('/api/v1/merchants')
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: MerchantApiItem[] }
    return body.data.map(mapMerchant)
  }

  async getById(id: string): Promise<MerchantAccount> {
    const res = await apiFetch(`/api/v1/merchants/${id}`)
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: MerchantApiItem }
    return mapMerchant(body.data)
  }

  async create(input: CreateMerchantInput): Promise<MerchantAccount> {
    const res = await apiFetch('/api/v1/merchants', {
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
    const body = (await res.json()) as { data: MerchantApiItem }
    return mapMerchant(body.data)
  }

  async update(input: UpdateMerchantInput): Promise<MerchantAccount> {
    const res = await apiFetch(`/api/v1/merchants/${input.id}`, {
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
    const body = (await res.json()) as { data: MerchantApiItem }
    return mapMerchant(body.data)
  }

  async remove(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/merchants/${id}`, { method: 'DELETE' })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
  }
}
