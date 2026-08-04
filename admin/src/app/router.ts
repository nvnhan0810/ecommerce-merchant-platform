import { createRouter, createWebHistory } from 'vue-router'
import AppShell from '@/app/AppShell.vue'
import OverviewPanel from '@/modules/dashboard/presentation/OverviewPanel.vue'
import MerchantList from '@/modules/merchants/presentation/MerchantList.vue'
import MerchantCreate from '@/modules/merchants/presentation/MerchantCreate.vue'
import MerchantEdit from '@/modules/merchants/presentation/MerchantEdit.vue'
import MerchantShow from '@/modules/merchants/presentation/MerchantShow.vue'
import UserList from '@/modules/users/presentation/UserList.vue'
import ProductList from '@/modules/products/presentation/ProductList.vue'
import ProductCreate from '@/modules/products/presentation/ProductCreate.vue'
import ProductEdit from '@/modules/products/presentation/ProductEdit.vue'
import ProductShow from '@/modules/products/presentation/ProductShow.vue'
import CategoryList from '@/modules/categories/presentation/CategoryList.vue'
import CategoryCreate from '@/modules/categories/presentation/CategoryCreate.vue'
import CategoryEdit from '@/modules/categories/presentation/CategoryEdit.vue'
import CategoryShow from '@/modules/categories/presentation/CategoryShow.vue'
import OrderList from '@/modules/orders/presentation/OrderList.vue'
import OrderShow from '@/modules/orders/presentation/OrderShow.vue'
import DeliverySimulator from '@/modules/orders/presentation/DeliverySimulator.vue'
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
        { path: 'merchants/new', component: MerchantCreate },
        { path: 'merchants/:id/edit', component: MerchantEdit },
        { path: 'merchants/:id', component: MerchantShow },
        { path: 'categories', component: CategoryList },
        { path: 'categories/new', component: CategoryCreate },
        { path: 'categories/:id/edit', component: CategoryEdit },
        { path: 'categories/:id', component: CategoryShow },
        { path: 'products', component: ProductList },
        { path: 'products/new', component: ProductCreate },
        { path: 'products/:id/edit', component: ProductEdit },
        { path: 'products/:id', component: ProductShow },
        { path: 'orders', component: OrderList },
        { path: 'orders/:id', component: OrderShow },
        { path: 'delivery-simulator', component: DeliverySimulator },
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
