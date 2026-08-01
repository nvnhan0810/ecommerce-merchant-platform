export class MerchantStats {
  constructor(
    readonly productCount: number,
    readonly orderCount: number,
    readonly revenueCents: number,
  ) {}

  get revenueLabel(): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
      maximumFractionDigits: 0,
    }).format(this.revenueCents)
  }
}

export interface DashboardRepository {
  loadStats(): Promise<MerchantStats>
}
