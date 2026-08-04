<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import type { Product } from '../domain/product'

const props = defineProps<{
  initial?: Product | null
  submitLabel?: string
  pending?: boolean
  error?: string
  showRemoveImage?: boolean
}>()

const emit = defineEmits<{
  submit: [
    payload: {
      name: string
      description: string
      priceCents: number
      currency: string
      stock: number
      file: File | null
    },
  ]
  cancel: []
  removeImage: []
}>()

const form = reactive({
  name: '',
  description: '',
  priceCents: 0,
  currency: 'VND',
  stock: 0,
})

const pendingFile = ref<File | null>(null)
const localPreview = ref('')
const existingImageUrl = ref('')

function clearLocalPreview(): void {
  if (localPreview.value) {
    URL.revokeObjectURL(localPreview.value)
  }
  localPreview.value = ''
}

function hydrate(): void {
  clearLocalPreview()
  pendingFile.value = null
  if (props.initial) {
    form.name = props.initial.name
    form.description = props.initial.description
    form.priceCents = props.initial.priceCents
    form.currency = props.initial.currency
    form.stock = props.initial.stock
    existingImageUrl.value = props.initial.imageUrl
  } else {
    form.name = ''
    form.description = ''
    form.priceCents = 0
    form.currency = 'VND'
    form.stock = 0
    existingImageUrl.value = ''
  }
}

watch(
  () => props.initial,
  () => hydrate(),
  { immediate: true, deep: true },
)

onBeforeUnmount(() => clearLocalPreview())

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  clearLocalPreview()
  pendingFile.value = file
  if (file) {
    localPreview.value = URL.createObjectURL(file)
  }
}

function onSubmit(): void {
  emit('submit', {
    name: form.name.trim(),
    description: form.description.trim(),
    priceCents: Number(form.priceCents),
    currency: form.currency.trim() || 'VND',
    stock: Number(form.stock),
    file: pendingFile.value,
  })
}

const previewSrc = () => localPreview.value || existingImageUrl.value
</script>

<template>
  <form class="form" @submit.prevent="onSubmit" aria-label="Product form">
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
        Giá (VND)
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
    <label>
      Hình ảnh
      <input type="file" accept="image/jpeg,image/png,image/webp,image/gif" @change="onFileChange" />
    </label>
    <div v-if="previewSrc()" class="image-preview">
      <img :src="previewSrc()" alt="Product preview" />
      <button
        v-if="showRemoveImage && existingImageUrl && !pendingFile"
        type="button"
        class="ghost"
        @click="emit('removeImage')"
      >
        Xóa ảnh
      </button>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="actions">
      <button type="submit" class="primary" :disabled="pending">
        {{ pending ? 'Đang lưu…' : submitLabel || 'Lưu' }}
      </button>
      <button type="button" class="ghost" @click="emit('cancel')">Hủy</button>
    </div>
  </form>
</template>

<style scoped>
.form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 720px;
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
textarea {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.55rem 0.7rem;
  font: inherit;
  font-weight: 400;
}

.image-preview {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.image-preview img {
  width: 96px;
  height: 96px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
}

.actions {
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
