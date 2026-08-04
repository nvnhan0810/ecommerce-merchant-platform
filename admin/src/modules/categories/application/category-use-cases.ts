import type {
  CategoryRepository,
  CreateCategoryInput,
  UpdateCategoryInput,
  CategoryStatus,
} from '../domain/category'

export class ListCategoriesUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(status?: string) {
    return this.repo.list(status)
  }
}

export class GetCategoryUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(id: string) {
    return this.repo.getById(id)
  }
}

export class CreateCategoryUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(input: CreateCategoryInput) {
    return this.repo.create(input)
  }
}

export class UpdateCategoryUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(input: UpdateCategoryInput) {
    return this.repo.update(input)
  }
}

export class UpdateCategoryStatusUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(id: string, status: CategoryStatus) {
    return this.repo.updateStatus(id, status)
  }
}

export class DeleteCategoryUseCase {
  constructor(private readonly repo: CategoryRepository) {}
  execute(id: string) {
    return this.repo.remove(id)
  }
}
