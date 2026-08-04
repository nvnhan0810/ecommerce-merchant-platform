import { Merchant, type MerchantRepository } from '../domain/merchant'

type MerchantApiItem = {
  id: string
  display_name: string
  avatar_url?: string
  country_code?: string
  province_code?: string
  ward_code?: string
  country_name?: string
  province_name?: string
  ward_name?: string
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpMerchantRepository implements MerchantRepository {
  async getById(id: string): Promise<Merchant> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/merchants/${id}`)
    if (!res.ok) {
      throw new Error(`Failed to load merchant (${res.status})`)
    }
    const body = (await res.json()) as { data: MerchantApiItem }
    const item = body.data
    return new Merchant(
      item.id,
      item.display_name,
      item.avatar_url ?? '',
      item.country_code ?? '',
      item.province_code ?? '',
      item.ward_code ?? '',
      item.country_name ?? '',
      item.province_name ?? '',
      item.ward_name ?? '',
    )
  }
}
