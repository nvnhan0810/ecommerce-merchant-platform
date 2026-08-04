export class Merchant {
  constructor(
    readonly id: string,
    readonly displayName: string,
    readonly avatarUrl: string = '',
    readonly countryCode: string = '',
    readonly provinceCode: string = '',
    readonly wardCode: string = '',
    readonly countryName: string = '',
    readonly provinceName: string = '',
    readonly wardName: string = '',
  ) {
    if (!id.trim()) {
      throw new Error('Merchant id is required')
    }
  }

  get formattedLocation(): string {
    return [this.wardName, this.provinceName].filter(Boolean).join(', ')
  }
}

export interface MerchantRepository {
  getById(id: string): Promise<Merchant>
}
