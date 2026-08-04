export interface UserAddress {
  id: string
  userId: string
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
  countryName?: string
  provinceName?: string
  wardName?: string
  latitude?: number | null
  longitude?: number | null
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export interface AddressInput {
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
  latitude?: number | null
  longitude?: number | null
  isDefault: boolean
}

export interface Country {
  code: string
  name: string
  nameEn: string
  isDefault: boolean
}

export interface Province {
  code: string
  countryCode: string
  name: string
  nameEn: string
  latitude?: number | null
  longitude?: number | null
}

export interface Ward {
  code: string
  provinceCode: string
  name: string
  nameEn: string
  latitude?: number | null
  longitude?: number | null
}

export interface AddressRepository {
  list(): Promise<UserAddress[]>
  get(id: string): Promise<UserAddress>
  create(input: AddressInput): Promise<UserAddress>
  update(id: string, input: AddressInput): Promise<UserAddress>
  delete(id: string): Promise<void>
}

export interface GeoRepository {
  listCountries(): Promise<Country[]>
  listProvinces(countryCode?: string): Promise<Province[]>
  listWards(provinceCode: string): Promise<Ward[]>
}

export function formatFullAddress(addr: UserAddress): string {
  return [addr.addressLine, addr.wardName, addr.provinceName, addr.countryName]
    .map((p) => (p ?? '').trim())
    .filter(Boolean)
    .join(', ')
}
