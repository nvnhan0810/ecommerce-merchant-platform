import type { AuthRepository, AuthSession, SessionStore } from '../domain/session'

export class LoginUseCase {
  constructor(
    private readonly repo: AuthRepository,
    private readonly store: SessionStore,
  ) {}

  async execute(email: string, password: string): Promise<AuthSession> {
    const session = await this.repo.login(email, password)
    this.store.save(session)
    return session
  }
}

export class GetStoredSessionUseCase {
  constructor(private readonly store: SessionStore) {}

  execute(): AuthSession | null {
    return this.store.load()
  }
}

export class LogoutUseCase {
  constructor(private readonly store: SessionStore) {}

  execute(): void {
    this.store.clear()
  }
}

export class UpdateProfileUseCase {
  constructor(
    private readonly repo: AuthRepository,
    private readonly store: SessionStore,
  ) {}

  async execute(input: { email: string; displayName: string; password?: string }): Promise<AuthSession> {
    const current = this.store.load()
    if (!current) {
      throw new Error('Chưa đăng nhập')
    }
    const session = await this.repo.updateProfile(current.accessToken, input)
    this.store.save(session)
    return session
  }
}
