import { describe, expect, it } from 'vitest'
import { MerchantAccount } from './merchant'

describe('MerchantAccount', () => {
  it('should_default_role_to_merchant', () => {
    const m = new MerchantAccount('1', 'a@b.c', 'Shop')
    expect(m.role).toBe('merchant')
  })
})
