<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import { ListPaymentCallbacksUseCase } from '../application/payment-callback-use-cases'
import { HttpPaymentCallbackRepository } from '../infrastructure/http-payment-callback-repository'
import {
  PAYMENT_CALLBACK_CHANNEL_OPTIONS,
  PAYMENT_PROVIDER_OPTIONS,
} from '../domain/payment-callback'

const listCallbacks = new ListPaymentCallbacksUseCase(new HttpPaymentCallbackRepository())

const providerFilter = ref('')
const channelFilter = ref('')
const refInput = ref('')
const appliedProvider = ref('')
const appliedChannel = ref('')
const appliedRef = ref('')

const filterKey = computed(() => ({
  provider: appliedProvider.value,
  channel: appliedChannel.value,
  merchTxnRef: appliedRef.value,
}))

const { data: items, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'payment-callbacks', filterKey.value]),
  queryFn: () =>
    listCallbacks.execute({
      provider: filterKey.value.provider || undefined,
      channel: filterKey.value.channel || undefined,
      merchTxnRef: filterKey.value.merchTxnRef || undefined,
    }),
})

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('vi-VN')
  } catch {
    return iso
  }
}

function applyFilters(): void {
  appliedProvider.value = providerFilter.value
  appliedChannel.value = channelFilter.value
  appliedRef.value = refInput.value.trim()
}

function clearFilters(): void {
  providerFilter.value = ''
  channelFilter.value = ''
  refInput.value = ''
  appliedProvider.value = ''
  appliedChannel.value = ''
  appliedRef.value = ''
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <h1>Payment callbacks</h1>
        <p class="subtitle">Nhật ký IPN / Return từ các cổng thanh toán</p>
      </div>
      <RouterLink class="ghost" to="/payments">Cài đặt Payments</RouterLink>
    </header>

    <form class="filters" @submit.prevent="applyFilters">
      <label>
        Provider
        <select v-model="providerFilter">
          <option value="">Tất cả</option>
          <option v-for="opt in PAYMENT_PROVIDER_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </label>
      <label>
        Channel
        <select v-model="channelFilter">
          <option value="">Tất cả</option>
          <option v-for="opt in PAYMENT_CALLBACK_CHANNEL_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </label>
      <label>
        Merch txn ref
        <input v-model="refInput" type="search" placeholder="vpc_MerchTxnRef…" />
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
    <ul v-else-if="items" class="list" aria-label="Payment callbacks">
      <li v-for="item in items" :key="item.id">
        <RouterLink class="item" :to="`/payment-callbacks/${item.id}`">
          <div class="row">
            <strong class="provider">{{ item.providerLabel }}</strong>
            <span class="pill" :data-channel="item.channel">{{ item.channelLabel }}</span>
            <span class="pill" :data-ok="item.success ? '1' : '0'">
              {{ item.success ? (item.paid ? 'Paid' : 'Failed txn') : 'Error' }}
            </span>
            <span v-if="item.httpMethod" class="muted method">{{ item.httpMethod }}</span>
          </div>
          <p class="meta">
            <span class="code">{{ item.merchTxnRef || '—' }}</span>
            <span v-if="item.responseCode">· code {{ item.responseCode }}</span>
            <span v-if="item.paymentMethodLabel">· {{ item.paymentMethodLabel }}</span>
          </p>
          <p v-if="item.message || item.errorMessage" class="msg">
            {{ item.errorMessage || item.message }}
          </p>
          <p class="muted">{{ formatDate(item.createdAt) }}</p>
        </RouterLink>
        <RouterLink class="ghost" :to="`/payment-callbacks/${item.id}`">Chi tiết</RouterLink>
      </li>
      <li v-if="items.length === 0" class="empty">Chưa có callback nào. Thử thanh toán sandbox rồi kiểm tra lại.</li>
    </ul>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.header h1 {
  margin: 0;
}

.subtitle {
  margin: 0.35rem 0 0;
  color: #64748b;
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
  min-width: 11rem;
  background: #fff;
}

.filter-actions {
  display: flex;
  gap: 0.5rem;
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

.item {
  display: block;
  text-decoration: none;
  color: inherit;
  min-width: 0;
  flex: 1;
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  align-items: center;
}

.provider {
  font-weight: 700;
}

.pill {
  display: inline-flex;
  align-items: center;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 650;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #475569;
}

.pill[data-channel='ipn'] {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

.pill[data-channel='return'] {
  background: #f5f3ff;
  color: #6d28d9;
  border-color: #ddd6fe;
}

.pill[data-ok='1'] {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

.pill[data-ok='0'] {
  background: #fef2f2;
  color: #b91c1c;
  border-color: #fecaca;
}

.code {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.02em;
}

.meta,
.msg,
.muted {
  margin: 0.3rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.msg {
  color: #334155;
}

.method {
  font-size: 0.78rem;
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
