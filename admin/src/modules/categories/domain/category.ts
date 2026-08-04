export type CategoryStatus = 'pending' | 'approved' | 'rejected'

export class Category {
  constructor(
    readonly id: string,
    readonly name: string,
    readonly status: CategoryStatus,
    readonly statusLabel: string,
    readonly createdByMerchantId: string = '',
    readonly createdAt: string = '',
  ) {}

  get isApproved(): boolean {
    return this.status === 'approved'
  }

  get isPending(): boolean {
    return this.status === 'pending'
  }
}

export type CreateCategoryInput = {
  name: string
}

export type UpdateCategoryInput = {
  id: string
  name: string
}

export interface CategoryRepository {
  list(status?: string): Promise<Category[]>
  getById(id: string): Promise<Category>
  create(input: CreateCategoryInput): Promise<Category>
  update(input: UpdateCategoryInput): Promise<Category>
  updateStatus(id: string, status: CategoryStatus): Promise<Category>
  remove(id: string): Promise<void>
}
