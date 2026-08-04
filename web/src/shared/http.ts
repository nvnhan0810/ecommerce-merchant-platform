import { GetStoredSessionUseCase, LogoutUseCase } from '@/modules/auth/application/auth-use-cases'
import { LocalStorageSessionStore } from '@/modules/auth/infrastructure/local-storage-session-store'

const store = new LocalStorageSessionStore()

export function getAccessToken(): string | null {
  return new GetStoredSessionUseCase(store).execute()?.accessToken ?? null
}

export function clearSession(): void {
  new LogoutUseCase(store).execute()
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const base = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
  const headers = new Headers(init.headers)
  const token = getAccessToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }
  const res = await fetch(`${base}${path}`, { ...init, headers })
  if (res.status === 401) {
    clearSession()
  }
  return res
}
