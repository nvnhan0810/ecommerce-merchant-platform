export class UserAccount {
  constructor(
    readonly id: string,
    readonly email: string,
    readonly displayName: string,
    readonly role: string = 'user',
  ) {}
}

export type CreateUserInput = {
  email: string
  displayName: string
  password: string
}

export type UpdateUserInput = {
  id: string
  email: string
  displayName: string
  password?: string
}

export interface UserRepository {
  list(): Promise<UserAccount[]>
  getById(id: string): Promise<UserAccount>
  create(input: CreateUserInput): Promise<UserAccount>
  update(input: UpdateUserInput): Promise<UserAccount>
  remove(id: string): Promise<void>
}
