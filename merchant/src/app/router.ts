import { createRouter, createWebHistory } from 'vue-router'
import DashboardPanel from '@/modules/dashboard/presentation/DashboardPanel.vue'
import ProductList from '@/modules/products/presentation/ProductList.vue'
import ProductCreate from '@/modules/products/presentation/ProductCreate.vue'
import ProductEdit from '@/modules/products/presentation/ProductEdit.vue'
import ProductShow from '@/modules/products/presentation/ProductShow.vue'
import OrderList from '@/modules/orders/presentation/OrderList.vue'
import OrderShow from '@/modules/orders/presentation/OrderShow.vue'
import AppShell from '@/app/AppShell.vue'
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
        { path: '', component: DashboardPanel },
        { path: 'profile', component: () => import('@/modules/profile/presentation/MerchantProfilePage.vue') },
        { path: 'products', component: ProductList },
        { path: 'products/new', component: ProductCreate },
        { path: 'products/:id/edit', component: ProductEdit },
        { path: 'products/:id', component: ProductShow },
        { path: 'orders', component: OrderList },
        { path: 'orders/:id', component: OrderShow },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const session = new GetStoredSessionUseCase(sessionStore).execute()
  const isAuthed = Boolean(session?.isMerchant)

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
