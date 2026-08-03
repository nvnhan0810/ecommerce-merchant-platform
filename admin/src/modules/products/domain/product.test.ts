import { describe, expect, it } from 'vitest'
import { Product } from './product'

describe('Product', () => {
  it('should_keep_merchant_id', () => {
    const p = new Product('1', 'm1', 'Áo', 'desc', 10000, 'VND', 5)
    expect(p.merchantId).toBe('m1')
    expect(p.priceCents).toBe(10000)
  })
})
