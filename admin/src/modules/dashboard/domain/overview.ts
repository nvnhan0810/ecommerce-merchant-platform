export class AdminOverview {
  constructor(
    readonly userCount: number,
    readonly merchantCount: number,
  ) {}

  get totalAccounts(): number {
    return this.userCount + this.merchantCount
  }
}

export class MerchantAccount {
  constructor(
    readonly id: string,
    readonly email: string,
    readonly displayName: string,
  ) {}
}

export interface AdminRepository {
  loadOverview(): Promise<AdminOverview>
  listMerchants(): Promise<MerchantAccount[]>
}
