import { apiFetch } from '@/shared/http'

export type Country = { code: string; name: string; nameEn: string; isDefault: boolean }
export type Province = { code: string; countryCode: string; name: string }
export type Ward = { code: string; provinceCode: string; name: string }

export async function fetchCountries(): Promise<Country[]> {
  const res = await apiFetch('/api/v1/countries')
  if (!res.ok) throw new Error(`Failed to load countries (${res.status})`)
  const body = (await res.json()) as {
    data: Array<{ code: string; name: string; name_en: string; is_default: boolean }>
  }
  return body.data.map((c) => ({
    code: c.code,
    name: c.name,
    nameEn: c.name_en,
    isDefault: c.is_default,
  }))
}

export async function fetchProvinces(countryCode: string): Promise<Province[]> {
  if (!countryCode) return []
  const res = await apiFetch(`/api/v1/provinces?country_code=${encodeURIComponent(countryCode)}`)
  if (!res.ok) throw new Error(`Failed to load provinces (${res.status})`)
  const body = (await res.json()) as {
    data: Array<{ code: string; country_code: string; name: string }>
  }
  return body.data.map((p) => ({
    code: p.code,
    countryCode: p.country_code,
    name: p.name,
  }))
}

export async function fetchWards(provinceCode: string): Promise<Ward[]> {
  if (!provinceCode) return []
  const res = await apiFetch(`/api/v1/wards?province_code=${encodeURIComponent(provinceCode)}`)
  if (!res.ok) throw new Error(`Failed to load wards (${res.status})`)
  const body = (await res.json()) as {
    data: Array<{ code: string; province_code: string; name: string }>
  }
  return body.data.map((w) => ({
    code: w.code,
    provinceCode: w.province_code,
    name: w.name,
  }))
}
