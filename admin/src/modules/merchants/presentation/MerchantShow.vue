<script setup lang="ts">
import { computed } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'
import { DeleteMerchantUseCase, GetMerchantUseCase } from '../application/merchant-use-cases'
import { HttpMerchantRepository } from '../infrastructure/http-merchant-repository'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpMerchantRepository()
const getMerchant = new GetMerchantUseCase(repo)
const deleteMerchant = new DeleteMerchantUseCase(repo)

const merchantId = computed(() => String(route.params.id))

const { data: merchant, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'merchants', merchantId.value]),
  queryFn: () => getMerchant.execute(merchantId.value),
})

const deleteMutation = useMutation({
  mutationFn: () => deleteMerchant.execute(merchantId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
    await router.push('/merchants')
  },
})

async function onDelete(): Promise<void> {
  if (!merchant.value) return
  if (!window.confirm(`Xóa merchant "${merchant.value.displayName}"?`)) return
  await deleteMutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/merchants">← Merchants</RouterLink>
        <h1>{{ merchant?.displayName || 'Chi tiết merchant' }}</h1>
      </div>
      <div v-if="merchant" class="actions">
        <RouterLink class="ghost" :to="`/merchants/${merchant.id}/edit`">Sửa</RouterLink>
        <button
          type="button"
          class="danger"
          :disabled="deleteMutation.isPending.value"
          @click="onDelete"
        >
          Xóa
        </button>
      </div>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <article v-else-if="merchant" class="detail">
      <div class="avatar" aria-hidden="true">
        <img v-if="merchant.avatarUrl" :src="merchant.avatarUrl" alt="" />
        <span v-else>{{ merchant.displayName.charAt(0) }}</span>
      </div>
      <dl>
        <div>
          <dt>Tên hiển thị</dt>
          <dd>{{ merchant.displayName }}</dd>
        </div>
        <div>
          <dt>Email</dt>
          <dd>{{ merchant.email }}</dd>
        </div>
        <div>
          <dt>Địa chỉ</dt>
          <dd>{{ merchant.formattedAddress || '—' }}</dd>
        </div>
        <div>
          <dt>Role</dt>
          <dd>{{ merchant.role }}</dd>
        </div>
      </dl>
    </article>
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
  gap: 1rem;
  align-items: flex-start;
  flex-wrap: wrap;
}

.header h1 {
  margin: 0.35rem 0 0;
}

.back {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9rem;
}

.back:hover {
  color: #0f172a;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.detail {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.1rem;
  display: grid;
  grid-template-columns: 96px 1fr;
  gap: 1.25rem;
  align-items: start;
}

.avatar {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
  background: #f1f5f9;
  overflow: hidden;
  display: grid;
  place-items: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: #64748b;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

dl {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

dt {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #64748b;
  margin-bottom: 0.2rem;
}

dd {
  margin: 0;
  font-size: 1rem;
  color: #0f172a;
}

a.ghost,
button {
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

.error {
  color: #b91c1c;
}

@media (max-width: 800px) {
  .detail {
    grid-template-columns: 1fr;
  }
}
</style>
