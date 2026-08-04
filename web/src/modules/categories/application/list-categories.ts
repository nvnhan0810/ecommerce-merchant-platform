import type { CategoryRepository } from '../domain/category'

export class ListCategoriesUseCase {
  constructor(private readonly repo: CategoryRepository) {}

  execute() {
    return this.repo.listApproved()
  }
}
