import { describe, expect, it } from 'vitest'
import { Money } from './product'

describe('Money', () => {
  it('should_format_vnd_currency', () => {
    const money = new Money(149000, 'VND')
    expect(money.format()).toMatch(/149/)
  })
})
