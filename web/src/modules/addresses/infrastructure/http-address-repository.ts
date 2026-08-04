import type {
  AddressInput,
  AddressRepository,
  Country,
  GeoRepository,
  Province,
  UserAddress,
  Ward,
} from '../domain/address'
import { apiFetch } from '@/shared/http'

type AddressApiItem = {
  id: string
  user_id: string
  address_line: string
  country_code: string
  province_code: string
  ward_code: string
  country_name?: string
  province_name?: string
  ward_name?: string
  latitude?: number | null
  longitude?: number | null
  is_default: boolean
  created_at: string
  updated_at: string
}

function mapAddress(item: AddressApiItem): UserAddress {
  return {
    id: item.id,
    userId: item.user_id,
    addressLine: item.address_line,
    countryCode: item.country_code || 'VN',
    provinceCode: item.province_code || '',
    wardCode: item.ward_code || '',
    countryName: item.country_name,
    provinceName: item.province_name,
    wardName: item.ward_name,
    latitude: item.latitude,
    longitude: item.longitude,
    isDefault: item.is_default,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  }
}

function toBody(input: AddressInput) {
  return {
    address_line: input.addressLine,
    country_code: input.countryCode || 'VN',
    province_code: input.provinceCode,
    ward_code: input.wardCode,
    latitude: input.latitude ?? null,
    longitude: input.longitude ?? null,
    is_default: input.isDefault,
  }
}

export class HttpAddressRepository implements AddressRepository {
  async list(): Promise<UserAddress[]> {
    const res = await apiFetch('/api/v1/me/addresses')
    if (!res.ok) {
      throw new Error('Không thể lấy danh sách địa chỉ')
    }
    const body = (await res.json()) as { data: AddressApiItem[] }
    return body.data.map(mapAddress)
  }

  async get(id: string): Promise<UserAddress> {
    const res = await apiFetch(`/api/v1/me/addresses/${id}`)
    if (!res.ok) {
      throw new Error('Không thể lấy địa chỉ')
    }
    const body = (await res.json()) as { data: AddressApiItem }
    return mapAddress(body.data)
  }

  async create(input: AddressInput): Promise<UserAddress> {
    const res = await apiFetch('/api/v1/me/addresses', {
      method: 'POST',
      body: JSON.stringify(toBody(input)),
    })
    if (!res.ok) {
      throw new Error('Không thể thêm địa chỉ')
    }
    const body = (await res.json()) as { data: AddressApiItem }
    return mapAddress(body.data)
  }

  async update(id: string, input: AddressInput): Promise<UserAddress> {
    const res = await apiFetch(`/api/v1/me/addresses/${id}`, {
      method: 'PUT',
      body: JSON.stringify(toBody(input)),
    })
    if (!res.ok) {
      throw new Error('Không thể cập nhật địa chỉ')
    }
    const body = (await res.json()) as { data: AddressApiItem }
    return mapAddress(body.data)
  }

  async delete(id: string): Promise<void> {
    const res = await apiFetch(`/api/v1/me/addresses/${id}`, {
      method: 'DELETE',
    })
    if (!res.ok) {
      throw new Error('Không thể xoá địa chỉ')
    }
  }
}

export class HttpGeoRepository implements GeoRepository {
  async listCountries(): Promise<Country[]> {
    const res = await apiFetch('/api/v1/countries')
    if (!res.ok) {
      throw new Error('Không thể lấy danh sách quốc gia')
    }
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

  async listProvinces(countryCode = 'VN'): Promise<Province[]> {
    const res = await apiFetch(`/api/v1/provinces?country_code=${encodeURIComponent(countryCode)}`)
    if (!res.ok) {
      throw new Error('Không thể lấy danh sách tỉnh/thành')
    }
    const body = (await res.json()) as {
      data: Array<{
        code: string
        country_code: string
        name: string
        name_en: string
        latitude?: number | null
        longitude?: number | null
      }>
    }
    return body.data.map((p) => ({
      code: p.code,
      countryCode: p.country_code,
      name: p.name,
      nameEn: p.name_en,
      latitude: p.latitude,
      longitude: p.longitude,
    }))
  }

  async listWards(provinceCode: string): Promise<Ward[]> {
    const res = await apiFetch(`/api/v1/wards?province_code=${encodeURIComponent(provinceCode)}`)
    if (!res.ok) {
      throw new Error('Không thể lấy danh sách phường/xã')
    }
    const body = (await res.json()) as {
      data: Array<{
        code: string
        province_code: string
        name: string
        name_en: string
        latitude?: number | null
        longitude?: number | null
      }>
    }
    return body.data.map((w) => ({
      code: w.code,
      provinceCode: w.province_code,
      name: w.name,
      nameEn: w.name_en,
      latitude: w.latitude,
      longitude: w.longitude,
    }))
  }
}
