import { describe, expect, it } from 'vitest'
import { Product } from './product'

describe('Product', () => {
  it('should_keep_merchant_and_image', () => {
    const p = new Product('1', 'm1', 'Áo', 'desc', 10000, 'VND', 5, 'products/1/a.jpg', 'https://api/x')
    expect(p.merchantId).toBe('m1')
    expect(p.imageKey).toBe('products/1/a.jpg')
    expect(p.imageUrl).toBe('https://api/x')
  })
})
