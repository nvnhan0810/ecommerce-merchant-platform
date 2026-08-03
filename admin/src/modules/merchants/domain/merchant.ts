export class MerchantAccount {
  constructor(
    readonly id: string,
    readonly email: string,
    readonly displayName: string,
    readonly role: string = 'merchant',
  ) {}
}

export type CreateMerchantInput = {
  email: string
  displayName: string
  password: string
}

export type UpdateMerchantInput = {
  id: string
  email: string
  displayName: string
  password?: string
}

export interface MerchantRepository {
  list(): Promise<MerchantAccount[]>
  getById(id: string): Promise<MerchantAccount>
  create(input: CreateMerchantInput): Promise<MerchantAccount>
  update(input: UpdateMerchantInput): Promise<MerchantAccount>
  remove(id: string): Promise<void>
}
