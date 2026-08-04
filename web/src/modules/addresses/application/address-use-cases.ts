import type {
  AddressInput,
  AddressRepository,
  Country,
  GeoRepository,
  Province,
  UserAddress,
  Ward,
} from '../domain/address'

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

export class ListCountriesUseCase {
  constructor(private readonly repo: GeoRepository) {}
  execute(): Promise<Country[]> {
    return this.repo.listCountries()
  }
}

export class ListProvincesUseCase {
  constructor(private readonly repo: GeoRepository) {}
  execute(countryCode = 'VN'): Promise<Province[]> {
    return this.repo.listProvinces(countryCode)
  }
}

export class ListWardsUseCase {
  constructor(private readonly repo: GeoRepository) {}
  execute(provinceCode: string): Promise<Ward[]> {
    return this.repo.listWards(provinceCode)
  }
}
