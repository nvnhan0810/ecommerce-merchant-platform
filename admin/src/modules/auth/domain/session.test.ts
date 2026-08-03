import { describe, expect, it } from 'vitest'
import { AuthSession } from './session'

describe('AuthSession', () => {
  it('should_mark_admin_role', () => {
    const session = new AuthSession('tok', '1', 'a@b.c', 'Admin', 'admin')
    expect(session.isAdmin).toBe(true)
  })
})
