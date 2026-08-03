import { describe, expect, it } from 'vitest'
import { UserAccount } from './user'

describe('UserAccount', () => {
  it('should_default_role_to_user', () => {
    const u = new UserAccount('1', 'a@b.c', 'Buyer')
    expect(u.role).toBe('user')
  })
})
