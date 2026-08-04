import {
  MerchantAccount,
  type CreateMerchantInput,
  type MerchantRepository,
  type UpdateMerchantInput,
} from '../domain/merchant'
import { apiFetch, getAccessToken } from '@/shared/http'

type MerchantApiItem = {
  id: string
  email: string
  display_name: string
  role: string
  avatar_url?: string
  address_line?: string
  country_code?: string
  province_code?: string
  ward_code?: string
  country_name?: string
  province_name?: string
  ward_name?: string
}

function mapMerchant(item: MerchantApiItem): MerchantAccount {
  return new MerchantAccount(
    item.id,
    item.email,
    item.display_name,
    item.role,
    item.avatar_url ?? '',
    item.address_line ?? '',
    item.country_code ?? '',
    item.province_code ?? '',
    item.ward_code ?? '',
    item.country_name ?? '',
    item.province_name ?? '',
    item.ward_name ?? '',
  )
}

function addressBody(input: {
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
}) {
  return {
    address_line: input.addressLine,
    country_code: input.countryCode,
    province_code: input.provinceCode,
    ward_code: input.wardCode,
  }
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

function apiBase(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
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
        ...addressBody(input),
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
        ...addressBody(input),
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

  async uploadAvatar(id: string, file: File): Promise<MerchantAccount> {
    const form = new FormData()
    form.append('file', file)
    const headers = new Headers()
    const token = getAccessToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
    const res = await fetch(`${apiBase()}/api/v1/merchants/${id}/avatar`, {
      method: 'POST',
      headers,
      body: form,
    })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: MerchantApiItem }
    return mapMerchant(body.data)
  }

  async deleteAvatar(id: string): Promise<MerchantAccount> {
    const res = await apiFetch(`/api/v1/merchants/${id}/avatar`, { method: 'DELETE' })
    if (!res.ok) {
      throw new Error(await readError(res))
    }
    const body = (await res.json()) as { data: MerchantApiItem }
    return mapMerchant(body.data)
  }
}
