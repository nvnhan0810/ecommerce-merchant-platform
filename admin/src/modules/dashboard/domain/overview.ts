import { MerchantAccount } from '../../merchants/domain/merchant'

export class AdminOverview {
  constructor(
    readonly userCount: number,
    readonly merchantCount: number,
  ) {}

  get totalAccounts(): number {
    return this.userCount + this.merchantCount
  }
}

export { MerchantAccount }

export interface AdminRepository {
  loadOverview(): Promise<AdminOverview>
  listMerchants(): Promise<MerchantAccount[]>
}
