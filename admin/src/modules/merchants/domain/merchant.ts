export class MerchantAccount {
  constructor(
    readonly id: string,
    readonly email: string,
    readonly displayName: string,
    readonly role: string = 'merchant',
    readonly avatarUrl: string = '',
    readonly addressLine: string = '',
    readonly countryCode: string = '',
    readonly provinceCode: string = '',
    readonly wardCode: string = '',
    readonly countryName: string = '',
    readonly provinceName: string = '',
    readonly wardName: string = '',
  ) {}

  get formattedAddress(): string {
    const parts = [this.addressLine, this.wardName, this.provinceName, this.countryName].filter(Boolean)
    return parts.join(', ')
  }
}

export type MerchantAddressInput = {
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
}

export type CreateMerchantInput = {
  email: string
  displayName: string
  password: string
} & MerchantAddressInput

export type UpdateMerchantInput = {
  id: string
  email: string
  displayName: string
  password?: string
} & MerchantAddressInput

export interface MerchantRepository {
  list(): Promise<MerchantAccount[]>
  getById(id: string): Promise<MerchantAccount>
  create(input: CreateMerchantInput): Promise<MerchantAccount>
  update(input: UpdateMerchantInput): Promise<MerchantAccount>
  remove(id: string): Promise<void>
  uploadAvatar(id: string, file: File): Promise<MerchantAccount>
  deleteAvatar(id: string): Promise<MerchantAccount>
}
