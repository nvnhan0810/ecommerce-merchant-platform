import { createRouter, createWebHistory } from 'vue-router'
import AppShell from '@/app/AppShell.vue'
import OverviewPanel from '@/modules/dashboard/presentation/OverviewPanel.vue'
import MerchantList from '@/modules/merchants/presentation/MerchantList.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', component: OverviewPanel },
        { path: 'merchants', component: MerchantList },
      ],
    },
  ],
})
