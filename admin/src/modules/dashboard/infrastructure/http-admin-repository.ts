import {
  AdminOverview,
  MerchantAccount,
  type AdminRepository,
} from '../../dashboard/domain/overview'

type UserApiItem = {
  id: string
  email: string
  display_name: string
  role: string
}

function apiBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || ''
}

export class HttpAdminRepository implements AdminRepository {
  async loadOverview(): Promise<AdminOverview> {
    const [usersRes, merchantsRes] = await Promise.all([
      fetch(`${apiBaseUrl()}/api/v1/users`),
      fetch(`${apiBaseUrl()}/api/v1/merchants`),
    ])
    if (!usersRes.ok || !merchantsRes.ok) {
      throw new Error('Failed to load admin overview')
    }
    const users = (await usersRes.json()) as { data: UserApiItem[] }
    const merchants = (await merchantsRes.json()) as { data: UserApiItem[] }
    return new AdminOverview(users.data.length, merchants.data.length)
  }

  async listMerchants(): Promise<MerchantAccount[]> {
    const res = await fetch(`${apiBaseUrl()}/api/v1/merchants`)
    if (!res.ok) {
      throw new Error(`Failed to load merchants (${res.status})`)
    }
    const body = (await res.json()) as { data: UserApiItem[] }
    return body.data.map(
      (item) => new MerchantAccount(item.id, item.email, item.display_name),
    )
  }
}
