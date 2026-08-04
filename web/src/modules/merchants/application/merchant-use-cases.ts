import type { Merchant, MerchantRepository } from '../domain/merchant'

export class GetMerchantUseCase {
  constructor(private readonly repo: MerchantRepository) {}

  execute(id: string): Promise<Merchant> {
    return this.repo.getById(id)
  }
}
