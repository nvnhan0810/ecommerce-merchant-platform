import { AuthSession, type SessionStore } from '../domain/session'

const STORAGE_KEY = 'ecomerce.merchant.session'

type StoredSession = {
  accessToken: string
  userId: string
  email: string
  displayName: string
  role: string
}

export class LocalStorageSessionStore implements SessionStore {
  load(): AuthSession | null {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return null
    }
    try {
      const parsed = JSON.parse(raw) as StoredSession
      return new AuthSession(
        parsed.accessToken,
        parsed.userId,
        parsed.email,
        parsed.displayName,
        parsed.role,
      )
    } catch {
      localStorage.removeItem(STORAGE_KEY)
      return null
    }
  }

  save(session: AuthSession): void {
    const payload: StoredSession = {
      accessToken: session.accessToken,
      userId: session.userId,
      email: session.email,
      displayName: session.displayName,
      role: session.role,
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
  }

  clear(): void {
    localStorage.removeItem(STORAGE_KEY)
  }
}
