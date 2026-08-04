import type { AddressInput, AddressRepository, UserAddress } from '../domain/address'
import { apiFetch } from '@/shared/http'

type AddressApiItem = {
  ID: string
  UserID: string
  RecipientName: string
  PhoneNumber: string
  AddressLine: string
  IsDefault: boolean
  CreatedAt: string
  UpdatedAt: string
}

function mapAddress(item: AddressApiItem): UserAddress {
  return {
    id: item.ID,
    userId: item.UserID,
    recipientName: item.RecipientName,
    phoneNumber: item.PhoneNumber,
    addressLine: item.AddressLine,
    isDefault: item.IsDefault,
    createdAt: item.CreatedAt,
    updatedAt: item.UpdatedAt,
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
      body: JSON.stringify({
        recipient_name: input.recipientName,
        phone_number: input.phoneNumber,
        address_line: input.addressLine,
        is_default: input.isDefault,
      }),
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
      body: JSON.stringify({
        recipient_name: input.recipientName,
        phone_number: input.phoneNumber,
        address_line: input.addressLine,
        is_default: input.isDefault,
      }),
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
