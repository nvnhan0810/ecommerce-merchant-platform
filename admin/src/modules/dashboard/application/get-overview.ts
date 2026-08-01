import type { AdminOverview, AdminRepository } from '../domain/overview'

export class GetAdminOverviewUseCase {
  constructor(private readonly repo: AdminRepository) {}

  execute(): Promise<AdminOverview> {
    return this.repo.loadOverview()
  }
}
