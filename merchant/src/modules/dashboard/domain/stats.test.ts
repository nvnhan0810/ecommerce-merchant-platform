import { describe, expect, it } from 'vitest'
import { MerchantStats } from './stats'

describe('MerchantStats', () => {
  it('should_format_revenue_label', () => {
    const stats = new MerchantStats(3, 0, 149000)
    expect(stats.revenueLabel).toMatch(/149/)
  })
})
