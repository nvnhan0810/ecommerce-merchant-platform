export class Category {
  constructor(
    readonly id: string,
    readonly name: string,
  ) {}
}

export interface CategoryRepository {
  listApproved(): Promise<Category[]>
}
