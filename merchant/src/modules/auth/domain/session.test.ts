import { describe, expect, it } from 'vitest'
import { AuthSession } from './session'

describe('AuthSession', () => {
  it('should_mark_merchant_role', () => {
    const session = new AuthSession('tok', '1', 'shop@ecomerce.local', 'Shop', 'merchant')
    expect(session.isMerchant).toBe(true)
  })

  it('should_reject_empty_token', () => {
    expect(() => new AuthSession('', '1', 'a@b.c', 'X', 'merchant')).toThrow()
  })
})
