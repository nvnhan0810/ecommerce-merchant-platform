import { describe, expect, it } from 'vitest'
import { AdminOverview } from './overview'

describe('AdminOverview', () => {
  it('should_sum_total_accounts', () => {
    const overview = new AdminOverview(4, 2)
    expect(overview.totalAccounts).toBe(6)
  })
})
