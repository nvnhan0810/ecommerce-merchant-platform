<script setup lang="ts">
import { computed, ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'
import {
  DeleteProductImageUseCase,
  GetProductUseCase,
  UpdateProductUseCase,
  UploadProductImageUseCase,
} from '../application/product-use-cases'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'
import ProductForm from './ProductForm.vue'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpProductRepository()
const getProduct = new GetProductUseCase(repo)
const updateProduct = new UpdateProductUseCase(repo)
const uploadImage = new UploadProductImageUseCase(repo)
const deleteImage = new DeleteProductImageUseCase(repo)
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())
const formError = ref('')

const productId = computed(() => String(route.params.id))

const { data: product, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'products', productId.value]),
  queryFn: () => getProduct.execute(productId.value),
})

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const saveMutation = useMutation({
  mutationFn: async (payload: {
    merchantId: string
    name: string
    description: string
    priceCents: number
    currency: string
    stock: number
    file: File | null
  }) => {
    if (!payload.merchantId) {
      throw new Error('Vui lòng chọn merchant')
    }
    let saved = await updateProduct.execute({
      id: productId.value,
      merchantId: payload.merchantId,
      name: payload.name,
      description: payload.description,
      priceCents: payload.priceCents,
      currency: payload.currency,
      stock: payload.stock,
    })
    if (payload.file) {
      saved = await uploadImage.execute(saved.id, payload.file)
    }
    return saved
  },
  onSuccess: async (saved) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
    await router.push(`/products/${saved.id}`)
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const removeImageMutation = useMutation({
  mutationFn: () => deleteImage.execute(productId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
    await refetch()
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

async function onRemoveImage(): Promise<void> {
  if (!window.confirm('Xóa ảnh sản phẩm?')) {
    return
  }
  await removeImageMutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" :to="`/products/${productId}`">← Chi tiết</RouterLink>
        <h1>Sửa sản phẩm</h1>
      </div>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ProductForm
      v-else-if="product"
      :merchants="merchants ?? []"
      :initial="product"
      submit-label="Cập nhật"
      show-remove-image
      :pending="saveMutation.isPending.value || removeImageMutation.isPending.value"
      :error="formError"
      @submit="(p) => { formError = ''; saveMutation.mutate(p) }"
      @cancel="router.push(`/products/${productId}`)"
      @remove-image="onRemoveImage"
    />
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

.error {
  color: #b91c1c;
}

.ghost {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #334155;
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
  cursor: pointer;
}
</style>
