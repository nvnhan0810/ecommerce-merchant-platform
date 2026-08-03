import type {
  CreateMerchantInput,
  MerchantAccount,
  MerchantRepository,
  UpdateMerchantInput,
} from '../domain/merchant'

export class ListMerchantsUseCase {
  constructor(private readonly repo: MerchantRepository) {}

  execute(): Promise<MerchantAccount[]> {
    return this.repo.list()
  }
}

export class CreateMerchantUseCase {
  constructor(private readonly repo: MerchantRepository) {}

  execute(input: CreateMerchantInput): Promise<MerchantAccount> {
    return this.repo.create(input)
  }
}

export class UpdateMerchantUseCase {
  constructor(private readonly repo: MerchantRepository) {}

  execute(input: UpdateMerchantInput): Promise<MerchantAccount> {
    return this.repo.update(input)
  }
}

export class DeleteMerchantUseCase {
  constructor(private readonly repo: MerchantRepository) {}

  execute(id: string): Promise<void> {
    return this.repo.remove(id)
  }
}
