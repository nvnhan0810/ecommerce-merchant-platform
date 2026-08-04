import type {
  MerchantProfile,
  MerchantProfileRepository,
  UpdateMerchantProfileInput,
} from '../domain/profile'

export class GetMerchantProfileUseCase {
  constructor(private readonly repo: MerchantProfileRepository) {}
  execute(): Promise<MerchantProfile> {
    return this.repo.getMe()
  }
}

export class UpdateMerchantProfileUseCase {
  constructor(private readonly repo: MerchantProfileRepository) {}
  execute(input: UpdateMerchantProfileInput): Promise<MerchantProfile> {
    return this.repo.update(input)
  }
}

export class UploadMerchantAvatarUseCase {
  constructor(private readonly repo: MerchantProfileRepository) {}
  execute(file: File): Promise<MerchantProfile> {
    return this.repo.uploadAvatar(file)
  }
}

export class DeleteMerchantAvatarUseCase {
  constructor(private readonly repo: MerchantProfileRepository) {}
  execute(): Promise<MerchantProfile> {
    return this.repo.deleteAvatar()
  }
}
