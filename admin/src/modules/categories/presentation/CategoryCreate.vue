<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { useRouter, RouterLink } from 'vue-router'
import { CreateCategoryUseCase } from '../application/category-use-cases'
import { HttpCategoryRepository } from '../infrastructure/http-category-repository'

const router = useRouter()
const queryClient = useQueryClient()
const createCategory = new CreateCategoryUseCase(new HttpCategoryRepository())
const formError = ref('')
const form = reactive({ name: '' })

const saveMutation = useMutation({
  mutationFn: () => createCategory.execute({ name: form.name.trim() }),
  onSuccess: async (cat) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] })
    await router.push(`/categories/${cat.id}`)
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
        <RouterLink class="back" to="/categories">← Danh mục</RouterLink>
        <h1>Tạo danh mục</h1>
      </div>
    </header>
    <form class="form" @submit.prevent="() => { formError = ''; saveMutation.mutate() }">
      <label>
        Tên danh mục
        <input v-model="form.name" required maxlength="120" />
      </label>
      <p class="hint">Danh mục do admin tạo sẽ được duyệt ngay và hiện trên storefront.</p>
      <p v-if="formError" class="error" role="alert">{{ formError }}</p>
      <div class="actions">
        <button type="submit" class="primary" :disabled="saveMutation.isPending.value">
          {{ saveMutation.isPending.value ? 'Đang lưu…' : 'Tạo danh mục' }}
        </button>
        <button type="button" class="ghost" @click="router.push('/categories')">Hủy</button>
      </div>
    </form>
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

.form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 520px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
}

input {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.55rem 0.7rem;
  font: inherit;
  font-weight: 400;
}

.hint {
  margin: 0;
  color: #64748b;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

button {
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
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

.error {
  color: #b91c1c;
  margin: 0;
}
</style>
