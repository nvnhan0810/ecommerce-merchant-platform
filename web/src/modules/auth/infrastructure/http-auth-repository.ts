import { AuthSession, type AuthRepository } from '../domain/session'

type LoginResponse = {
  access_token: string
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

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) return body.error
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`
}

export class HttpAuthRepository implements AuthRepository {
  async login(email: string, password: string): Promise<AuthSession> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/auth/user/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
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
    const res = await fetch(`${apiBaseUrl()}/api/v1/auth/user/me`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    if (!res.ok) {
      throw new Error(await readError(res))
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

  async updateProfile(
    accessToken: string,
    input: { email: string; displayName: string; password?: string },
  ): Promise<AuthSession> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/auth/user/me`, {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: input.email,
        display_name: input.displayName,
        password: input.password ?? '',
      }),
    })
    if (!res.ok) {
      throw new Error(await readError(res))
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
