export class CartItem {
  constructor(
    readonly productId: string,
    readonly name: string,
    readonly unitPriceCents: number,
    readonly quantity: number,
  ) {
    if (quantity < 1) {
      throw new Error('quantity must be >= 1')
    }
  }

  get lineTotalCents(): number {
    return this.unitPriceCents * this.quantity
  }
}

export class Cart {
  constructor(readonly items: CartItem[] = []) {}

  get totalCents(): number {
    return this.items.reduce((sum, item) => sum + item.lineTotalCents, 0)
  }

  get itemCount(): number {
    return this.items.reduce((sum, item) => sum + item.quantity, 0)
  }
}
