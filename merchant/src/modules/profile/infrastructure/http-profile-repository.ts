import {
  MerchantProfile,
  type MerchantProfileRepository,
  type UpdateMerchantProfileInput,
} from '../domain/profile'
import { apiFetch, getAccessToken } from '@/shared/http'

type ProfileApiItem = {
  id: string
  email: string
  display_name: string
  avatar_url?: string
  address_line?: string
  country_code?: string
  province_code?: string
  ward_code?: string
  country_name?: string
  province_name?: string
  ward_name?: string
}

function mapProfile(item: ProfileApiItem): MerchantProfile {
  return new MerchantProfile(
    item.id,
    item.email,
    item.display_name,
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

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

function apiBase(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpMerchantProfileRepository implements MerchantProfileRepository {
  async getMe(): Promise<MerchantProfile> {
    const res = await apiFetch('/api/v1/auth/merchant/me')
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: ProfileApiItem }
    return mapProfile(body.data)
  }

  async update(input: UpdateMerchantProfileInput): Promise<MerchantProfile> {
    const res = await apiFetch('/api/v1/auth/merchant/me', {
      method: 'PUT',
      body: JSON.stringify({
        display_name: input.displayName,
        password: input.password ?? '',
        address_line: input.addressLine,
        country_code: input.countryCode,
        province_code: input.provinceCode,
        ward_code: input.wardCode,
      }),
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: ProfileApiItem }
    return mapProfile(body.data)
  }

  async uploadAvatar(file: File): Promise<MerchantProfile> {
    const form = new FormData()
    form.append('file', file)
    const headers = new Headers()
    const token = getAccessToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
    const res = await fetch(`${apiBase()}/api/v1/auth/merchant/me/avatar`, {
      method: 'POST',
      headers,
      body: form,
    })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: ProfileApiItem }
    return mapProfile(body.data)
  }

  async deleteAvatar(): Promise<MerchantProfile> {
    const res = await apiFetch('/api/v1/auth/merchant/me/avatar', { method: 'DELETE' })
    if (!res.ok) throw new Error(await readError(res))
    const body = (await res.json()) as { data: ProfileApiItem }
    return mapProfile(body.data)
  }
}
