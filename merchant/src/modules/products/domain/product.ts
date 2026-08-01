export class Money {
  constructor(
    readonly amountCents: number,
    readonly currency: string,
  ) {}

  format(): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: this.currency || 'VND',
      maximumFractionDigits: 0,
    }).format(this.amountCents)
  }
}

export class ProductId {
  constructor(readonly value: string) {
    if (!value.trim()) {
      throw new Error('ProductId is required')
    }
  }
}

export class Product {
  constructor(
    readonly id: ProductId,
    readonly name: string,
    readonly description: string,
    readonly price: Money,
    readonly stock: number,
  ) {}
}

export interface MerchantProductRepository {
  list(): Promise<Product[]>
}
