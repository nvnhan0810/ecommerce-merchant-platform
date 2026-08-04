<script setup lang="ts">
import { computed } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import {
  DeleteCategoryUseCase,
  GetCategoryUseCase,
  UpdateCategoryStatusUseCase,
} from '../application/category-use-cases'
import { HttpCategoryRepository } from '../infrastructure/http-category-repository'
import type { CategoryStatus } from '../domain/category'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpCategoryRepository()
const getCategory = new GetCategoryUseCase(repo)
const deleteCategory = new DeleteCategoryUseCase(repo)
const updateStatus = new UpdateCategoryStatusUseCase(repo)
const categoryId = computed(() => String(route.params.id))

const { data: category, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'categories', categoryId.value]),
  queryFn: () => getCategory.execute(categoryId.value),
})

const deleteMutation = useMutation({
  mutationFn: () => deleteCategory.execute(categoryId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
    await router.push('/categories')
  },
})

const statusMutation = useMutation({
  mutationFn: (status: CategoryStatus) => updateStatus.execute(categoryId.value, status),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
    await refetch()
  },
})

async function onDelete(): Promise<void> {
  if (!category.value) return
  if (!window.confirm(`Xóa danh mục "${category.value.name}"?`)) return
  await deleteMutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/categories">← Danh mục</RouterLink>
        <h1>{{ category?.name || 'Chi tiết danh mục' }}</h1>
      </div>
      <div v-if="category" class="actions">
        <button
          v-if="category.isPending"
          type="button"
          class="primary"
          :disabled="statusMutation.isPending.value"
          @click="statusMutation.mutate('approved')"
        >
          Duyệt
        </button>
        <button
          v-if="category.isPending"
          type="button"
          class="ghost"
          :disabled="statusMutation.isPending.value"
          @click="statusMutation.mutate('rejected')"
        >
          Từ chối
        </button>
        <RouterLink class="ghost" :to="`/categories/${category.id}/edit`">Sửa</RouterLink>
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
    <article v-else-if="category" class="detail">
      <dl>
        <div>
          <dt>Tên</dt>
          <dd>{{ category.name }}</dd>
        </div>
        <div>
          <dt>Trạng thái</dt>
          <dd>
            <span class="badge" :data-status="category.status">{{ category.statusLabel }}</span>
          </dd>
        </div>
        <div v-if="category.createdByMerchantId">
          <dt>Tạo bởi merchant</dt>
          <dd>{{ category.createdByMerchantId }}</dd>
        </div>
        <div v-if="category.createdAt">
          <dt>Ngày tạo</dt>
          <dd>{{ category.createdAt }}</dd>
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

.actions {
  display: flex;
  gap: 0.45rem;
  flex-wrap: wrap;
  align-items: flex-start;
}

.detail {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  max-width: 640px;
}

dl {
  margin: 0;
  display: grid;
  gap: 0.85rem;
}

dt {
  font-size: 0.8rem;
  color: #64748b;
  font-weight: 600;
}

dd {
  margin: 0.2rem 0 0;
}

.badge {
  display: inline-block;
  font-size: 0.85rem;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  background: #e2e8f0;
}

.badge[data-status='approved'] {
  background: #dcfce7;
  color: #166534;
}

.badge[data-status='pending'] {
  background: #fef3c7;
  color: #92400e;
}

.badge[data-status='rejected'] {
  background: #fee2e2;
  color: #991b1b;
}

button,
.ghost {
  border-radius: 8px;
  padding: 0.4rem 0.75rem;
  font: inherit;
  cursor: pointer;
  text-decoration: none;
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
  background: #fff;
  color: #b91c1c;
}

.error {
  color: #b91c1c;
}
</style>
