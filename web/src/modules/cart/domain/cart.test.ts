import { describe, expect, it } from 'vitest'
import { Cart, CartItem } from '../domain/cart'

describe('Cart', () => {
  it('should_add_and_sum_line_totals', () => {
    const cart = new Cart([]).add(new CartItem('p1', 'm1', 'Áo', 100000, 'VND', 2))
    expect(cart.itemCount).toBe(2)
    expect(cart.totalCents).toBe(200000)
  })
})
