import type { AddressInput, AddressRepository, UserAddress } from '../domain/address'

export class ListAddressesUseCase {
  constructor(private readonly repo: AddressRepository) {}
  execute(): Promise<UserAddress[]> {
    return this.repo.list()
  }
}

export class GetAddressUseCase {
  constructor(private readonly repo: AddressRepository) {}
  execute(id: string): Promise<UserAddress> {
    return this.repo.get(id)
  }
}

export class CreateAddressUseCase {
  constructor(private readonly repo: AddressRepository) {}
  execute(input: AddressInput): Promise<UserAddress> {
    return this.repo.create(input)
  }
}

export class UpdateAddressUseCase {
  constructor(private readonly repo: AddressRepository) {}
  execute(id: string, input: AddressInput): Promise<UserAddress> {
    return this.repo.update(id, input)
  }
}

export class DeleteAddressUseCase {
  constructor(private readonly repo: AddressRepository) {}
  execute(id: string): Promise<void> {
    return this.repo.delete(id)
  }
}
