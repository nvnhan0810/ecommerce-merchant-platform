import type { CategoryRepository } from '../domain/category'

export class ListAssignableCategoriesUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute() {
    return this.repo.listAssignable()
  }
}

export class CreateCategoryUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(name: string) {
    return this.repo.create(name)
  }
}
