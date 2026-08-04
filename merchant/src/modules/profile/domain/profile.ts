export class MerchantProfile {
  constructor(
    readonly id: string,
    readonly email: string,
    readonly displayName: string,
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

export type UpdateMerchantProfileInput = {
  displayName: string
  password?: string
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
}

export interface MerchantProfileRepository {
  getMe(): Promise<MerchantProfile>
  update(input: UpdateMerchantProfileInput): Promise<MerchantProfile>
  uploadAvatar(file: File): Promise<MerchantProfile>
  deleteAvatar(): Promise<MerchantProfile>
}
