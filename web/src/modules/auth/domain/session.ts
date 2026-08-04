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

  get isUser(): boolean {
    return this.role === 'user'
  }
}

export interface AuthRepository {
  login(email: string, password: string): Promise<AuthSession>
  me(accessToken: string): Promise<AuthSession>
  updateProfile(
    accessToken: string,
    input: { email: string; displayName: string; password?: string },
  ): Promise<AuthSession>
}

export interface SessionStore {
  load(): AuthSession | null
  save(session: AuthSession): void
  clear(): void
}
