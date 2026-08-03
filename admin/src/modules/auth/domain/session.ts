export class AuthSession {
  constructor(
    readonly accessToken: string,
    readonly userId: string,
    readonly email: string,
    readonly displayName: string,
    readonly role: string,
  ) {
    if (!accessToken.trim()) {
      throw new Error('accessToken is required')
    }
  }

  get isAdmin(): boolean {
    return this.role === 'admin'
  }
}

export interface AuthRepository {
  login(email: string, password: string): Promise<AuthSession>
  me(accessToken: string): Promise<AuthSession>
}

export interface SessionStore {
  load(): AuthSession | null
  save(session: AuthSession): void
  clear(): void
}
