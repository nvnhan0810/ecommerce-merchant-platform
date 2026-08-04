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
import {
  CreateCategoryUseCase,
  ListAssignableCategoriesUseCase,
} from '@/modules/categories/application/category-use-cases'
import { HttpCategoryRepository } from '@/modules/categories/infrastructure/http-category-repository'
import ProductForm from './ProductForm.vue'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpProductRepository()
const categoryRepo = new HttpCategoryRepository()
const getProduct = new GetProductUseCase(repo)
const updateProduct = new UpdateProductUseCase(repo)
const uploadImage = new UploadProductImageUseCase(repo)
const deleteImage = new DeleteProductImageUseCase(repo)
const listCategories = new ListAssignableCategoriesUseCase(categoryRepo)
const createCategory = new CreateCategoryUseCase(categoryRepo)
const formError = ref('')
const formRef = ref<{ selectCategory: (id: string) => void } | null>(null)

const productId = computed(() => String(route.params.id))

const { data: product, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['merchant', 'products', productId.value]),
  queryFn: () => getProduct.execute(productId.value),
})

const { data: categories, refetch: refetchCategories } = useQuery({
  queryKey: ['merchant', 'categories'],
  queryFn: () => listCategories.execute(),
})

const createCategoryMutation = useMutation({
  mutationFn: (name: string) => createCategory.execute(name),
  onSuccess: async (cat) => {
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'categories'] })
    await refetchCategories()
    formRef.value?.selectCategory(cat.id)
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const saveMutation = useMutation({
  mutationFn: async (payload: {
    name: string
    description: string
    priceCents: number
    currency: string
    stock: number
    categoryIds: string[]
    file: File | null
  }) => {
    let saved = await updateProduct.execute({
      id: productId.value,
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
  onSuccess: async (saved) => {
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'products'] })
    await router.push(`/products/${saved.id}`)
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const removeImageMutation = useMutation({
  mutationFn: () => deleteImage.execute(productId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'products'] })
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
      ref="formRef"
      :categories="categories ?? []"
      :initial="product"
      submit-label="Cập nhật"
      show-remove-image
      :pending="saveMutation.isPending.value || removeImageMutation.isPending.value"
      :creating-category="createCategoryMutation.isPending.value"
      :error="formError"
      @submit="(p) => { formError = ''; saveMutation.mutate(p) }"
      @cancel="router.push(`/products/${productId}`)"
      @remove-image="onRemoveImage"
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
