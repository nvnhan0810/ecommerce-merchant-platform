import type { DashboardRepository, MerchantStats } from '../domain/stats'

export class GetMerchantDashboardUseCase {
  constructor(private readonly repo: DashboardRepository) {}

  execute(): Promise<MerchantStats> {
    return this.repo.loadStats()
  }
}
