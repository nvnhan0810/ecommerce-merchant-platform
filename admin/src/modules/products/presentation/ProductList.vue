<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  CreateProductUseCase,
  DeleteProductUseCase,
  ListProductsUseCase,
  UpdateProductUseCase,
} from '../application/product-use-cases'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'
import type { Product } from '../domain/product'

const productRepo = new HttpProductRepository()
const listProducts = new ListProductsUseCase(productRepo)
const createProduct = new CreateProductUseCase(productRepo)
const updateProduct = new UpdateProductUseCase(productRepo)
const deleteProduct = new DeleteProductUseCase(productRepo)
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())

const queryClient = useQueryClient()
const formError = ref('')
const editingId = ref<string | null>(null)
const showForm = ref(false)

const form = reactive({
  merchantId: '',
  name: '',
  description: '',
  priceCents: 0,
  currency: 'VND',
  stock: 0,
})

const { data: products, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'products'],
  queryFn: () => listProducts.execute(),
})

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const merchantNameById = computed(() => {
  const map = new Map<string, string>()
  for (const m of merchants.value ?? []) {
    map.set(m.id, m.displayName || m.email)
  }
  return map
})

const isEditing = computed(() => editingId.value !== null)

function merchantLabel(merchantId: string): string {
  return merchantNameById.value.get(merchantId) ?? merchantId
}

function formatPrice(cents: number, currency: string): string {
  return `${cents.toLocaleString('vi-VN')} ${currency}`
}

function resetForm(): void {
  form.merchantId = merchants.value?.[0]?.id ?? ''
  form.name = ''
  form.description = ''
  form.priceCents = 0
  form.currency = 'VND'
  form.stock = 0
  editingId.value = null
  formError.value = ''
  showForm.value = false
}

function openCreate(): void {
  resetForm()
  form.merchantId = merchants.value?.[0]?.id ?? ''
  showForm.value = true
}

function openEdit(product: Product): void {
  editingId.value = product.id
  form.merchantId = product.merchantId
  form.name = product.name
  form.description = product.description
  form.priceCents = product.priceCents
  form.currency = product.currency
  form.stock = product.stock
  formError.value = ''
  showForm.value = true
}

const saveMutation = useMutation({
  mutationFn: async () => {
    const payload = {
      merchantId: form.merchantId,
      name: form.name.trim(),
      description: form.description.trim(),
      priceCents: Number(form.priceCents),
      currency: form.currency.trim() || 'VND',
      stock: Number(form.stock),
    }
    if (isEditing.value && editingId.value) {
      return updateProduct.execute({ id: editingId.value, ...payload })
    }
    return createProduct.execute(payload)
  },
  onSuccess: async () => {
    resetForm()
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteProduct.execute(id),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
  },
})

async function onSubmit(): Promise<void> {
  formError.value = ''
  if (!form.merchantId) {
    formError.value = 'Vui lòng chọn merchant'
    return
  }
  await saveMutation.mutateAsync()
}

async function onDelete(product: Product): Promise<void> {
  if (!window.confirm(`Xóa sản phẩm "${product.name}"?`)) {
    return
  }
  await deleteMutation.mutateAsync(product.id)
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Products</h1>
      <button type="button" class="primary" @click="openCreate">Thêm sản phẩm</button>
    </header>

    <form v-if="showForm" class="form" @submit.prevent="onSubmit" aria-label="Product form">
      <h2>{{ isEditing ? 'Sửa sản phẩm' : 'Tạo sản phẩm' }}</h2>
      <label>
        Merchant
        <select v-model="form.merchantId" required>
          <option disabled value="">Chọn merchant</option>
          <option v-for="m in merchants ?? []" :key="m.id" :value="m.id">
            {{ m.displayName }} ({{ m.email }})
          </option>
        </select>
      </label>
      <label>
        Tên sản phẩm
        <input v-model="form.name" required />
      </label>
      <label>
        Mô tả
        <textarea v-model="form.description" rows="3" />
      </label>
      <div class="row">
        <label>
          Giá (cents / VND)
          <input v-model.number="form.priceCents" type="number" min="1" required />
        </label>
        <label>
          Currency
          <input v-model="form.currency" required />
        </label>
        <label>
          Stock
          <input v-model.number="form.stock" type="number" min="0" required />
        </label>
      </div>
      <p v-if="formError" class="error" role="alert">{{ formError }}</p>
      <div class="actions">
        <button type="submit" class="primary" :disabled="saveMutation.isPending.value">
          {{ saveMutation.isPending.value ? 'Đang lưu…' : 'Lưu' }}
        </button>
        <button type="button" class="ghost" @click="resetForm">Hủy</button>
      </div>
    </form>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="products" class="list" aria-label="Danh sách sản phẩm">
      <li v-for="product in products" :key="product.id">
        <div>
          <strong>{{ product.name }}</strong>
          <p>{{ formatPrice(product.priceCents, product.currency) }} · stock {{ product.stock }}</p>
          <p class="merchant">Merchant: {{ merchantLabel(product.merchantId) }}</p>
        </div>
        <div class="row-actions">
          <button type="button" class="ghost" @click="openEdit(product)">Sửa</button>
          <button
            type="button"
            class="danger"
            :disabled="deleteMutation.isPending.value"
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

.form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form h2 {
  margin: 0;
  font-size: 1.05rem;
}

.row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
}

input,
select,
textarea {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.55rem 0.7rem;
  font: inherit;
  font-weight: 400;
}

.actions,
.row-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

button {
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
  cursor: pointer;
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

.merchant {
  color: #0f172a !important;
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

@media (max-width: 800px) {
  .row {
    grid-template-columns: 1fr;
  }
}
</style>
