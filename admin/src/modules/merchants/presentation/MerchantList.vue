<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import { DeleteMerchantUseCase, ListMerchantsUseCase } from '../application/merchant-use-cases'
import { HttpMerchantRepository } from '../infrastructure/http-merchant-repository'
import type { MerchantAccount } from '../domain/merchant'

const queryClient = useQueryClient()
const repo = new HttpMerchantRepository()
const listMerchants = new ListMerchantsUseCase(repo)
const deleteMerchant = new DeleteMerchantUseCase(repo)

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteMerchant.execute(id),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
  },
})

async function onDelete(merchant: MerchantAccount): Promise<void> {
  if (!window.confirm(`Xóa merchant "${merchant.displayName}"?`)) {
    return
  }
  await deleteMutation.mutateAsync(merchant.id)
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Merchants</h1>
      <RouterLink class="primary" to="/merchants/new">Thêm merchant</RouterLink>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="data" class="list" aria-label="Danh sách merchant">
      <li v-for="merchant in data" :key="merchant.id">
        <RouterLink class="item" :to="`/merchants/${merchant.id}`">
          <div class="avatar" aria-hidden="true">
            <img v-if="merchant.avatarUrl" :src="merchant.avatarUrl" alt="" />
            <span v-else>{{ merchant.displayName.charAt(0) }}</span>
          </div>
          <div>
            <strong>{{ merchant.displayName }}</strong>
            <p>{{ merchant.email }}</p>
            <p v-if="merchant.formattedAddress" class="addr">{{ merchant.formattedAddress }}</p>
          </div>
        </RouterLink>
        <div class="row-actions">
          <RouterLink class="ghost" :to="`/merchants/${merchant.id}`">Xem</RouterLink>
          <RouterLink class="ghost" :to="`/merchants/${merchant.id}/edit`">Sửa</RouterLink>
          <button
            type="button"
            class="danger"
            :disabled="deleteMutation.isPending.value"
            @click="onDelete(merchant)"
          >
            Xóa
          </button>
        </div>
      </li>
      <li v-if="data.length === 0" class="empty">Chưa có merchant nào.</li>
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
  align-items: center;
  gap: 1rem;
}

.header h1 {
  margin: 0;
}

.item {
  display: flex;
  gap: 0.85rem;
  align-items: center;
  text-decoration: none;
  color: inherit;
  min-width: 0;
  flex: 1;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
  background: #f1f5f9;
  overflow: hidden;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #64748b;
  flex-shrink: 0;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.addr {
  color: #64748b !important;
}

.row-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
}

a.primary,
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

button:disabled {
  opacity: 0.65;
  cursor: wait;
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

.danger {
  border: 1px solid #fecaca;
  background: #fff1f2;
  color: #b91c1c;
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
  margin: 0.2rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.empty {
  color: #64748b;
  justify-content: center;
}

.error {
  color: #b91c1c;
  margin: 0;
}
</style>
