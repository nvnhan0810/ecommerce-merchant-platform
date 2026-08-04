<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import { ref } from 'vue'
import {
  DeleteCategoryUseCase,
  ListCategoriesUseCase,
  UpdateCategoryStatusUseCase,
} from '../application/category-use-cases'
import { HttpCategoryRepository } from '../infrastructure/http-category-repository'
import type { Category, CategoryStatus } from '../domain/category'

const queryClient = useQueryClient()
const repo = new HttpCategoryRepository()
const listCategories = new ListCategoriesUseCase(repo)
const deleteCategory = new DeleteCategoryUseCase(repo)
const updateStatus = new UpdateCategoryStatusUseCase(repo)
const statusFilter = ref('')

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'categories', statusFilter],
  queryFn: () => listCategories.execute(statusFilter.value || undefined),
})

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteCategory.execute(id),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
  },
})

const statusMutation = useMutation({
  mutationFn: (payload: { id: string; status: CategoryStatus }) =>
    updateStatus.execute(payload.id, payload.status),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
  },
})

async function onDelete(cat: Category): Promise<void> {
  if (!window.confirm(`Xóa danh mục "${cat.name}"?`)) return
  await deleteMutation.mutateAsync(cat.id)
}

async function onApprove(cat: Category): Promise<void> {
  await statusMutation.mutateAsync({ id: cat.id, status: 'approved' })
}

async function onReject(cat: Category): Promise<void> {
  await statusMutation.mutateAsync({ id: cat.id, status: 'rejected' })
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Danh mục</h1>
      <RouterLink class="primary" to="/categories/new">Thêm danh mục</RouterLink>
    </header>

    <div class="filters">
      <label>
        Trạng thái
        <select v-model="statusFilter">
          <option value="">Tất cả</option>
          <option value="pending">Chờ duyệt</option>
          <option value="approved">Đã duyệt</option>
          <option value="rejected">Từ chối</option>
        </select>
      </label>
    </div>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="data" class="list" aria-label="Danh sách danh mục">
      <li v-for="cat in data" :key="cat.id">
        <RouterLink class="item" :to="`/categories/${cat.id}`">
          <div>
            <strong>{{ cat.name }}</strong>
            <p>
              <span class="badge" :data-status="cat.status">{{ cat.statusLabel }}</span>
              <span v-if="cat.createdByMerchantId" class="muted"> · do merchant tạo</span>
            </p>
          </div>
        </RouterLink>
        <div class="row-actions">
          <button
            v-if="cat.isPending"
            type="button"
            class="ghost"
            :disabled="statusMutation.isPending.value"
            @click="onApprove(cat)"
          >
            Duyệt
          </button>
          <button
            v-if="cat.isPending"
            type="button"
            class="ghost"
            :disabled="statusMutation.isPending.value"
            @click="onReject(cat)"
          >
            Từ chối
          </button>
          <RouterLink class="ghost" :to="`/categories/${cat.id}/edit`">Sửa</RouterLink>
          <button
            type="button"
            class="danger"
            :disabled="deleteMutation.isPending.value"
            @click="onDelete(cat)"
          >
            Xóa
          </button>
        </div>
      </li>
      <li v-if="data.length === 0" class="empty">Chưa có danh mục nào.</li>
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
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.header h1 {
  margin: 0;
}

.filters label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
  max-width: 220px;
}

select {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.45rem 0.7rem;
  font: inherit;
  font-weight: 400;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 0.75rem 1rem;
  flex-wrap: wrap;
}

.item {
  text-decoration: none;
  color: inherit;
  flex: 1;
  min-width: 200px;
}

.item p {
  margin: 0.25rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.badge {
  display: inline-block;
  font-size: 0.8rem;
  padding: 0.1rem 0.45rem;
  border-radius: 999px;
  background: #e2e8f0;
  color: #334155;
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

.muted {
  color: #94a3b8;
}

.row-actions {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.primary,
.ghost,
.danger {
  border-radius: 8px;
  padding: 0.4rem 0.75rem;
  font: inherit;
  text-decoration: none;
  cursor: pointer;
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

.empty {
  color: #64748b;
  justify-content: center;
}

.error {
  color: #b91c1c;
}
</style>
