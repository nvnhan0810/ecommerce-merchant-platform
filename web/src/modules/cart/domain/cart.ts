export class CartItem {
  constructor(
    readonly productId: string,
    readonly merchantId: string,
    readonly name: string,
    readonly unitPriceCents: number,
    readonly currency: string,
    readonly quantity: number,
    readonly imageUrl: string = '',
  ) {
    if (quantity < 1) {
      throw new Error('quantity must be >= 1')
    }
  }

  get lineTotalCents(): number {
    return this.unitPriceCents * this.quantity
  }

  withQuantity(quantity: number): CartItem {
    return new CartItem(
      this.productId,
      this.merchantId,
      this.name,
      this.unitPriceCents,
      this.currency,
      quantity,
      this.imageUrl,
    )
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

  add(item: CartItem): Cart {
    const existing = this.items.find((x) => x.productId === item.productId)
    if (!existing) {
      return new Cart([...this.items, item])
    }
    return new Cart(
      this.items.map((x) =>
        x.productId === item.productId ? x.withQuantity(x.quantity + item.quantity) : x,
      ),
    )
  }

  setQuantity(productId: string, quantity: number): Cart {
    if (quantity < 1) {
      return this.remove(productId)
    }
    return new Cart(
      this.items.map((x) => (x.productId === productId ? x.withQuantity(quantity) : x)),
    )
  }

  remove(productId: string): Cart {
    return new Cart(this.items.filter((x) => x.productId !== productId))
  }

  clear(): Cart {
    return new Cart([])
  }
}
