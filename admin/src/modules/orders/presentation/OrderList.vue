<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import { ListOrdersUseCase } from '../application/order-use-cases'
import { HttpOrderRepository } from '../infrastructure/http-order-repository'
import { ORDER_STATUS_OPTIONS } from '../domain/order'
import OrderStatusBadge from './OrderStatusBadge.vue'
import { ListUsersUseCase } from '@/modules/users/application/user-use-cases'
import { HttpUserRepository } from '@/modules/users/infrastructure/http-user-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'

const orderRepo = new HttpOrderRepository()
const listOrders = new ListOrdersUseCase(orderRepo)
const listUsers = new ListUsersUseCase(new HttpUserRepository())
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())

const codeInput = ref('')
const statusFilter = ref('')
const appliedCode = ref('')
const appliedStatus = ref('')

const filterKey = computed(() => ({
  code: appliedCode.value,
  status: appliedStatus.value,
}))

const { data: orders, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'orders', filterKey.value]),
  queryFn: () =>
    listOrders.execute({
      code: filterKey.value.code || undefined,
      status: filterKey.value.status || undefined,
    }),
})

const { data: users } = useQuery({
  queryKey: ['admin', 'users'],
  queryFn: () => listUsers.execute(),
})

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const userLabelById = computed(() => {
  const map = new Map<string, string>()
  for (const u of users.value ?? []) {
    map.set(u.id, u.displayName || u.email)
  }
  return map
})

const merchantLabelById = computed(() => {
  const map = new Map<string, string>()
  for (const m of merchants.value ?? []) {
    map.set(m.id, m.displayName || m.email)
  }
  return map
})

function userLabel(id: string): string {
  return userLabelById.value.get(id) ?? id
}

function merchantLabel(id: string): string {
  return merchantLabelById.value.get(id) ?? id
}

function formatMoney(cents: number, currency: string): string {
  return `${cents.toLocaleString('vi-VN')} ${currency}`
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('vi-VN')
  } catch {
    return iso
  }
}

function applyFilters(): void {
  appliedCode.value = codeInput.value.trim().toUpperCase()
  appliedStatus.value = statusFilter.value
}

function clearFilters(): void {
  codeInput.value = ''
  statusFilter.value = ''
  appliedCode.value = ''
  appliedStatus.value = ''
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Orders</h1>
    </header>

    <form class="filters" @submit.prevent="applyFilters">
      <label>
        Mã đơn
        <input v-model="codeInput" type="search" placeholder="VD: K7M2P9QX4A" maxlength="10" />
      </label>
      <label>
        Trạng thái
        <select v-model="statusFilter">
          <option value="">Tất cả</option>
          <option v-for="opt in ORDER_STATUS_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </label>
      <div class="filter-actions">
        <button type="submit" class="primary">Lọc</button>
        <button type="button" class="ghost" @click="clearFilters">Xóa lọc</button>
      </div>
    </form>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="orders" class="list" aria-label="Danh sách đơn hàng">
      <li v-for="order in orders" :key="order.id">
        <RouterLink class="item" :to="`/orders/${order.id}`">
          <div>
            <strong class="code">{{ order.code }}</strong>
            <p class="meta-line">
              <OrderStatusBadge :status="order.status" :label="order.statusLabel" />
              <span>· {{ formatMoney(order.totalCents, order.currency) }}</span>
            </p>
            <p>{{ userLabel(order.userId) }} → {{ merchantLabel(order.merchantId) }}</p>
            <p class="muted">{{ formatDate(order.createdAt) }}</p>
          </div>
        </RouterLink>
        <RouterLink class="ghost" :to="`/orders/${order.id}`">Xem</RouterLink>
      </li>
      <li v-if="orders.length === 0" class="empty">Không tìm thấy đơn hàng.</li>
    </ul>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header h1 {
  margin: 0;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.85rem;
  align-items: flex-end;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.85rem 1rem;
}

.filters label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.filters input,
.filters select {
  font: inherit;
  text-transform: none;
  letter-spacing: normal;
  color: #0f172a;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.45rem 0.65rem;
  min-width: 12rem;
  background: #fff;
}

.filter-actions {
  display: flex;
  gap: 0.5rem;
}

.item {
  display: block;
  text-decoration: none;
  color: inherit;
  min-width: 0;
  flex: 1;
}

.code {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.04em;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.muted {
  color: #94a3b8 !important;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.list li {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.85rem 1rem;
}

.list p {
  margin: 0.25rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.empty {
  color: #64748b;
  justify-content: center;
}

.error {
  color: #b91c1c;
}

button,
a.ghost {
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
}

.primary {
  border: 0;
  background: #0f172a;
  color: #fff;
}

.ghost {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #334155;
}
</style>
