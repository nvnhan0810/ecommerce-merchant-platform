import { createRouter, createWebHistory } from 'vue-router'
import DashboardPanel from '@/modules/dashboard/presentation/DashboardPanel.vue'
import ProductList from '@/modules/products/presentation/ProductList.vue'
import AppShell from '@/app/AppShell.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', component: DashboardPanel },
        { path: 'products', component: ProductList },
      ],
    },
  ],
})
