import type { AuthRepository, AuthSession, SessionStore } from '../domain/session'

export class LoginUseCase {
  constructor(
    private readonly repo: AuthRepository,
    private readonly store: SessionStore,
  ) {}

  async execute(email: string, password: string): Promise<AuthSession> {
    const session = await this.repo.login(email, password)
    if (!session.isMerchant) {
      throw new Error('merchant role required')
    }
    this.store.save(session)
    return session
  }
}

export class LogoutUseCase {
  constructor(private readonly store: SessionStore) {}

  execute(): void {
    this.store.clear()
  }
}

export class GetStoredSessionUseCase {
  constructor(private readonly store: SessionStore) {}

  execute(): AuthSession | null {
    return this.store.load()
  }
}
