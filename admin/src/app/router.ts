import { createRouter, createWebHistory } from 'vue-router'
import AppShell from '@/app/AppShell.vue'
import OverviewPanel from '@/modules/dashboard/presentation/OverviewPanel.vue'
import MerchantList from '@/modules/merchants/presentation/MerchantList.vue'
import UserList from '@/modules/users/presentation/UserList.vue'
import LoginPage from '@/modules/auth/presentation/LoginPage.vue'
import { GetStoredSessionUseCase } from '@/modules/auth/application/auth-use-cases'
import { LocalStorageSessionStore } from '@/modules/auth/infrastructure/local-storage-session-store'

const sessionStore = new LocalStorageSessionStore()

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      component: LoginPage,
      meta: { public: true },
    },
    {
      path: '/',
      component: AppShell,
      meta: { requiresAuth: true },
      children: [
        { path: '', component: OverviewPanel },
        { path: 'users', component: UserList },
        { path: 'merchants', component: MerchantList },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const session = new GetStoredSessionUseCase(sessionStore).execute()
  const isAuthed = Boolean(session?.isAdmin)

  if (to.meta.public) {
    if (isAuthed && to.path === '/login') {
      return '/'
    }
    return true
  }

  if (to.meta.requiresAuth && !isAuthed) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  return true
})
