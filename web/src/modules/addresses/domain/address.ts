export interface UserAddress {
  id: string
  userId: string
  recipientName: string
  phoneNumber: string
  addressLine: string
  isDefault: boolean
  createdAt: string
  updatedAt: string
}

export interface AddressInput {
  recipientName: string
  phoneNumber: string
  addressLine: string
  isDefault: boolean
}

export interface AddressRepository {
  list(): Promise<UserAddress[]>
  get(id: string): Promise<UserAddress>
  create(input: AddressInput): Promise<UserAddress>
  update(id: string, input: AddressInput): Promise<UserAddress>
  delete(id: string): Promise<void>
}
