<script setup lang="ts">
import { computed } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'
import { DeleteProductUseCase, GetProductUseCase } from '../application/product-use-cases'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpProductRepository()
const getProduct = new GetProductUseCase(repo)
const deleteProduct = new DeleteProductUseCase(repo)
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())

const productId = computed(() => String(route.params.id))

const { data: product, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'products', productId.value]),
  queryFn: () => getProduct.execute(productId.value),
})

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const merchantLabel = computed(() => {
  const id = product.value?.merchantId
  if (!id) return ''
  const m = (merchants.value ?? []).find((x) => x.id === id)
  return m ? `${m.displayName} (${m.email})` : id
})

function formatPrice(cents: number, currency: string): string {
  return `${cents.toLocaleString('vi-VN')} ${currency}`
}

const deleteMutation = useMutation({
  mutationFn: () => deleteProduct.execute(productId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
    await router.push('/products')
  },
})

async function onDelete(): Promise<void> {
  if (!product.value) return
  if (!window.confirm(`Xóa sản phẩm "${product.value.name}"?`)) return
  await deleteMutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/products">← Products</RouterLink>
        <h1>{{ product?.name || 'Chi tiết sản phẩm' }}</h1>
      </div>
      <div class="actions" v-if="product">
        <RouterLink class="ghost" :to="`/products/${product.id}/edit`">Sửa</RouterLink>
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
    <article v-else-if="product" class="detail">
      <img
        v-if="product.imageUrl"
        class="hero"
        :src="product.imageUrl"
        :alt="product.name"
      />
      <div v-else class="hero placeholder">Chưa có ảnh</div>
      <dl>
        <div>
          <dt>Merchant</dt>
          <dd>{{ merchantLabel }}</dd>
        </div>
        <div>
          <dt>Giá</dt>
          <dd>{{ formatPrice(product.priceCents, product.currency) }}</dd>
        </div>
        <div>
          <dt>Stock</dt>
          <dd>{{ product.stock }}</dd>
        </div>
        <div>
          <dt>Mô tả</dt>
          <dd>{{ product.description || '—' }}</dd>
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
  grid-template-columns: 240px 1fr;
  gap: 1.25rem;
}

.hero {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
}

.hero.placeholder {
  display: grid;
  place-items: center;
  color: #94a3b8;
  font-size: 0.9rem;
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
