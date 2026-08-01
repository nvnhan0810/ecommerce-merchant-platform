import type { MerchantAccount, AdminRepository } from '../../dashboard/domain/overview'

export class ListMerchantsUseCase {
  constructor(private readonly repo: AdminRepository) {}

  execute(): Promise<MerchantAccount[]> {
    return this.repo.listMerchants()
  }
}
