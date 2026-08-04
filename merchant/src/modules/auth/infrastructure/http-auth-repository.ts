import { AuthSession, type AuthRepository } from '../domain/session'

type LoginResponse = {
  access_token: string
  token_type: string
  user: {
    id: string
    email: string
    display_name: string
    role: string
  }
}

type MeResponse = {
  data: {
    id: string
    email: string
    display_name: string
    role: string
  }
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpAuthRepository implements AuthRepository {
  async login(email: string, password: string): Promise<AuthSession> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/auth/merchant/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      let message = `Login failed (${res.status})`
      try {
        const body = (await res.json()) as { error?: string }
        if (body.error) {
          message = body.error
        }
      } catch {
        // ignore
      }
      throw new Error(message)
    }
    const body = (await res.json()) as LoginResponse
    return new AuthSession(
      body.access_token,
      body.user.id,
      body.user.email,
      body.user.display_name,
      body.user.role,
    )
  }

  async me(accessToken: string): Promise<AuthSession> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/auth/merchant/me`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    if (!res.ok) {
      throw new Error(`Session expired (${res.status})`)
    }
    const body = (await res.json()) as MeResponse
    return new AuthSession(
      accessToken,
      body.data.id,
      body.data.email,
      body.data.display_name,
      body.data.role,
    )
  }
}
