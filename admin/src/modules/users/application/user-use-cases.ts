import type {
  CreateUserInput,
  UpdateUserInput,
  UserAccount,
  UserRepository,
} from '../domain/user'

export class ListUsersUseCase {
  constructor(private readonly repo: UserRepository) {}

  execute(): Promise<UserAccount[]> {
    return this.repo.list()
  }
}

export class CreateUserUseCase {
  constructor(private readonly repo: UserRepository) {}

  execute(input: CreateUserInput): Promise<UserAccount> {
    return this.repo.create(input)
  }
}

export class UpdateUserUseCase {
  constructor(private readonly repo: UserRepository) {}

  execute(input: UpdateUserInput): Promise<UserAccount> {
    return this.repo.update(input)
  }
}

export class DeleteUserUseCase {
  constructor(private readonly repo: UserRepository) {}

  execute(id: string): Promise<void> {
    return this.repo.remove(id)
  }
}
