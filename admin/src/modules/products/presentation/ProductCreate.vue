<script setup lang="ts">
import { ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import {
  CreateProductUseCase,
  UploadProductImageUseCase,
} from '../application/product-use-cases'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'
import {
  CreateCategoryUseCase,
  ListCategoriesUseCase,
} from '@/modules/categories/application/category-use-cases'
import { HttpCategoryRepository } from '@/modules/categories/infrastructure/http-category-repository'
import ProductForm from './ProductForm.vue'

const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpProductRepository()
const categoryRepo = new HttpCategoryRepository()
const createProduct = new CreateProductUseCase(repo)
const uploadImage = new UploadProductImageUseCase(repo)
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())
const listCategories = new ListCategoriesUseCase(categoryRepo)
const createCategory = new CreateCategoryUseCase(categoryRepo)
const formError = ref('')
const formRef = ref<{ selectCategory: (id: string) => void } | null>(null)

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const { data: categories, refetch: refetchCategories } = useQuery({
  queryKey: ['admin', 'categories'],
  queryFn: () => listCategories.execute(),
})

const createCategoryMutation = useMutation({
  mutationFn: (name: string) => createCategory.execute({ name }),
  onSuccess: async (cat) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
    await refetchCategories()
    formRef.value?.selectCategory(cat.id)
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const saveMutation = useMutation({
  mutationFn: async (payload: {
    merchantId: string
    name: string
    description: string
    priceCents: number
    currency: string
    stock: number
    categoryIds: string[]
    file: File | null
  }) => {
    if (!payload.merchantId) {
      throw new Error('Vui lòng chọn merchant')
    }
    let saved = await createProduct.execute({
      merchantId: payload.merchantId,
      name: payload.name,
      description: payload.description,
      priceCents: payload.priceCents,
      currency: payload.currency,
      stock: payload.stock,
      categoryIds: payload.categoryIds,
    })
    if (payload.file) {
      saved = await uploadImage.execute(saved.id, payload.file)
    }
    return saved
  },
  onSuccess: async (product) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'products'] })
    await router.push(`/products/${product.id}`)
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/products">← Products</RouterLink>
        <h1>Tạo sản phẩm</h1>
      </div>
    </header>
    <ProductForm
      ref="formRef"
      :merchants="merchants ?? []"
      :categories="categories ?? []"
      submit-label="Tạo sản phẩm"
      :pending="saveMutation.isPending.value"
      :creating-category="createCategoryMutation.isPending.value"
      :error="formError"
      @submit="(p) => { formError = ''; saveMutation.mutate(p) }"
      @cancel="router.push('/products')"
      @create-category="(name) => { formError = ''; createCategoryMutation.mutate(name) }"
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
</style>
