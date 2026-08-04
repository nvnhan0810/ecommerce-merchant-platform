<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { RouterLink } from 'vue-router'
import { DeleteProductUseCase, ListProductsUseCase } from '../application/product-use-cases'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import type { Product } from '../domain/product'

const queryClient = useQueryClient()
const productRepo = new HttpProductRepository()
const listProducts = new ListProductsUseCase(productRepo)
const deleteProduct = new DeleteProductUseCase(productRepo)

const { data: products, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['merchant', 'products'],
  queryFn: () => listProducts.execute(),
})

function formatPrice(cents: number, currency: string): string {
  return `${cents.toLocaleString('vi-VN')} ${currency}`
}

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteProduct.execute(id),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'products'] })
  },
})

async function onDelete(product: Product): Promise<void> {
  if (!product.canDelete) {
    window.alert('Không thể xóa sản phẩm đã có đơn hàng.')
    return
  }
  if (!window.confirm(`Xóa sản phẩm "${product.name}"?`)) {
    return
  }
  try {
    await deleteMutation.mutateAsync(product.id)
  } catch (e) {
    window.alert((e as Error).message)
  }
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Sản phẩm</h1>
      <RouterLink class="primary" to="/products/new">Thêm sản phẩm</RouterLink>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="products" class="list" aria-label="Danh sách sản phẩm">
      <li v-for="product in products" :key="product.id">
        <RouterLink class="item" :to="`/products/${product.id}`">
          <img
            v-if="product.imageUrl"
            class="thumb"
            :src="product.imageUrl"
            :alt="product.name"
          />
          <div v-else class="thumb placeholder" aria-hidden="true" />
          <div>
            <strong>{{ product.name }}</strong>
            <p>{{ formatPrice(product.priceCents, product.currency) }} · stock {{ product.stock }}</p>
            <p v-if="product.hasOrders" class="hint">Đã có đơn — không thể xóa</p>
          </div>
        </RouterLink>
        <div class="row-actions">
          <RouterLink class="ghost" :to="`/products/${product.id}`">Xem</RouterLink>
          <RouterLink class="ghost" :to="`/products/${product.id}/edit`">Sửa</RouterLink>
          <button
            type="button"
            class="danger"
            :disabled="!product.canDelete || deleteMutation.isPending.value"
            :title="product.canDelete ? 'Xóa sản phẩm' : 'Sản phẩm đã có đơn hàng'"
            @click="onDelete(product)"
          >
            Xóa
          </button>
        </div>
      </li>
      <li v-if="products.length === 0" class="empty">Chưa có sản phẩm nào.</li>
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

.thumb {
  width: 72px;
  height: 72px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  flex-shrink: 0;
}

.thumb.placeholder {
  display: inline-block;
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
  opacity: 0.55;
  cursor: not-allowed;
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

.hint {
  color: #b45309 !important;
  font-weight: 500;
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
