export type CategoryStatus = 'pending' | 'approved' | 'rejected'

export class Category {
  constructor(
    readonly id: string,
    readonly name: string,
    readonly status: CategoryStatus,
    readonly statusLabel: string,
  ) {}

  get isPending(): boolean {
    return this.status === 'pending'
  }
}

export interface CategoryRepository {
  listAssignable(): Promise<Category[]>
  create(name: string): Promise<Category>
}
