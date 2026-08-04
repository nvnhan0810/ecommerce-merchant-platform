import { AuthSession, type SessionStore } from '../domain/session'

const KEY = 'ecomerce.user.session'

type Stored = {
  accessToken: string
  userId: string
  email: string
  displayName: string
  role: string
}

export class LocalStorageSessionStore implements SessionStore {
  load(): AuthSession | null {
    try {
      const raw = localStorage.getItem(KEY)
      if (!raw) return null
      const data = JSON.parse(raw) as Stored
      return new AuthSession(data.accessToken, data.userId, data.email, data.displayName, data.role)
    } catch {
      return null
    }
  }

  save(session: AuthSession): void {
    const data: Stored = {
      accessToken: session.accessToken,
      userId: session.userId,
      email: session.email,
      displayName: session.displayName,
      role: session.role,
    }
    localStorage.setItem(KEY, JSON.stringify(data))
  }

  clear(): void {
    localStorage.removeItem(KEY)
  }
}
