import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useSyncExternalStore,
  type JSX,
  type ReactNode,
} from 'react'
import { AuthSession } from '../domain/session'
import {
  GetStoredSessionUseCase,
  LoginUseCase,
  LogoutUseCase,
  UpdateProfileUseCase,
} from '../application/auth-use-cases'
import { HttpAuthRepository } from '../infrastructure/http-auth-repository'
import { LocalStorageSessionStore } from '../infrastructure/local-storage-session-store'

const store = new LocalStorageSessionStore()
const repo = new HttpAuthRepository()
const getStored = new GetStoredSessionUseCase(store)
const loginUseCase = new LoginUseCase(repo, store)
const logoutUseCase = new LogoutUseCase(store)
const updateProfileUseCase = new UpdateProfileUseCase(repo, store)

type AuthContextValue = {
  session: AuthSession | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  updateProfile: (input: { email: string; displayName: string; password?: string }) => Promise<void>
  refresh: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

const listeners = new Set<() => void>()
function subscribe(cb: () => void): () => void {
  listeners.add(cb)
  return () => listeners.delete(cb)
}
function emit(): void {
  listeners.forEach((cb) => cb())
}

let cachedRaw: string | null | undefined = undefined
let cachedSession: AuthSession | null = null

function readSession(): AuthSession | null {
  const raw = localStorage.getItem('ecomerce.user.session')
  if (raw === cachedRaw) {
    return cachedSession
  }
  cachedRaw = raw
  cachedSession = getStored.execute()
  return cachedSession
}

function invalidateSessionCache(): void {
  cachedRaw = undefined
  cachedSession = null
}

function getSnapshot(): AuthSession | null {
  return readSession()
}

export function AuthProvider({ children }: { children: ReactNode }): JSX.Element {
  const session = useSyncExternalStore(subscribe, getSnapshot, () => null)

  const refresh = useCallback(() => {
    invalidateSessionCache()
    emit()
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      async login(email, password) {
        await loginUseCase.execute(email, password)
        refresh()
      },
      logout() {
        logoutUseCase.execute()
        refresh()
      },
      async updateProfile(input) {
        await updateProfileUseCase.execute(input)
        refresh()
      },
      refresh,
    }),
    [session, refresh],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
