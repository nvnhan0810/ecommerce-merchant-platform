<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import type { MerchantAccount } from '../domain/merchant'
import { fetchCountries, fetchProvinces, fetchWards } from '@/shared/geo'
import SearchableSelect from '@/shared/SearchableSelect.vue'

export type MerchantFormPayload = {
  email: string
  displayName: string
  password: string
  addressLine: string
  countryCode: string
  provinceCode: string
  wardCode: string
  file: File | null
  removeAvatar: boolean
}

const props = defineProps<{
  initial?: MerchantAccount | null
  submitLabel?: string
  pending?: boolean
  error?: string
}>()

const emit = defineEmits<{
  submit: [payload: MerchantFormPayload]
  cancel: []
}>()

const form = reactive({
  email: '',
  displayName: '',
  password: '',
  addressLine: '',
  countryCode: 'VN',
  provinceCode: '',
  wardCode: '',
})

const pendingFile = ref<File | null>(null)
const localPreview = ref('')
const existingAvatarUrl = ref('')
const markRemoveAvatar = ref(false)

const { data: countries } = useQuery({
  queryKey: ['geo', 'countries'],
  queryFn: fetchCountries,
})

const { data: provinces } = useQuery({
  queryKey: computed(() => ['geo', 'provinces', form.countryCode]),
  queryFn: () => fetchProvinces(form.countryCode),
  enabled: computed(() => Boolean(form.countryCode)),
})

const { data: wards } = useQuery({
  queryKey: computed(() => ['geo', 'wards', form.provinceCode]),
  queryFn: () => fetchWards(form.provinceCode),
  enabled: computed(() => Boolean(form.provinceCode)),
})

const countryOptions = computed(() =>
  (countries.value ?? []).map((c) => ({ value: c.code, label: c.name })),
)
const provinceOptions = computed(() =>
  (provinces.value ?? []).map((p) => ({ value: p.code, label: p.name })),
)
const wardOptions = computed(() =>
  (wards.value ?? []).map((w) => ({ value: w.code, label: w.name })),
)

function clearLocalPreview(): void {
  if (localPreview.value) {
    URL.revokeObjectURL(localPreview.value)
  }
  localPreview.value = ''
}

function hydrate(): void {
  clearLocalPreview()
  pendingFile.value = null
  markRemoveAvatar.value = false
  if (props.initial) {
    form.email = props.initial.email
    form.displayName = props.initial.displayName
    form.password = ''
    form.addressLine = props.initial.addressLine
    form.countryCode = props.initial.countryCode || 'VN'
    form.provinceCode = props.initial.provinceCode
    form.wardCode = props.initial.wardCode
    existingAvatarUrl.value = props.initial.avatarUrl
  } else {
    form.email = ''
    form.displayName = ''
    form.password = ''
    form.addressLine = ''
    form.countryCode = 'VN'
    form.provinceCode = ''
    form.wardCode = ''
    existingAvatarUrl.value = ''
  }
}

watch(
  () => props.initial,
  () => hydrate(),
  { immediate: true, deep: true },
)

onBeforeUnmount(() => clearLocalPreview())

function onCountryChange(code: string): void {
  form.countryCode = code || 'VN'
  form.provinceCode = ''
  form.wardCode = ''
}

function onProvinceChange(code: string): void {
  form.provinceCode = code
  form.wardCode = ''
}

function onWardChange(code: string): void {
  form.wardCode = code
}

function onAvatarChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  clearLocalPreview()
  pendingFile.value = file
  markRemoveAvatar.value = false
  if (file) {
    localPreview.value = URL.createObjectURL(file)
  }
}

function clearAvatar(): void {
  clearLocalPreview()
  pendingFile.value = null
  existingAvatarUrl.value = ''
  markRemoveAvatar.value = true
}

function onSubmit(): void {
  emit('submit', {
    email: form.email.trim(),
    displayName: form.displayName.trim(),
    password: form.password,
    addressLine: form.addressLine.trim(),
    countryCode: form.countryCode,
    provinceCode: form.provinceCode,
    wardCode: form.wardCode,
    file: pendingFile.value,
    removeAvatar: markRemoveAvatar.value,
  })
}

const previewSrc = computed(() => localPreview.value || existingAvatarUrl.value)
const isEditing = computed(() => Boolean(props.initial))
</script>

<template>
  <form class="form" aria-label="Merchant form" @submit.prevent="onSubmit">
    <div class="avatarRow">
      <div class="avatarPreview" aria-hidden="true">
        <img v-if="previewSrc" :src="previewSrc" alt="" />
        <span v-else>{{ form.displayName?.charAt(0) || '?' }}</span>
      </div>
      <div class="avatarActions">
        <label class="fileBtn">
          Chọn avatar
          <input type="file" accept="image/jpeg,image/png,image/webp,image/gif" @change="onAvatarChange" />
        </label>
        <button
          v-if="previewSrc"
          type="button"
          class="ghost"
          @click="clearAvatar"
        >
          Xóa avatar
        </button>
      </div>
    </div>

    <label>
      Tên hiển thị
      <input v-model="form.displayName" required />
    </label>
    <label>
      Email
      <input v-model="form.email" type="email" required autocomplete="off" />
    </label>
    <label>
      Password
      <input
        v-model="form.password"
        type="password"
        :required="!isEditing"
        :placeholder="isEditing ? 'Để trống nếu giữ password cũ' : ''"
        autocomplete="new-password"
      />
    </label>

    <h3>Địa chỉ gian hàng</h3>
    <label>
      Địa chỉ chi tiết
      <input v-model="form.addressLine" placeholder="Số nhà, đường…" />
    </label>
    <label>
      Quốc gia
      <SearchableSelect
        :options="countryOptions"
        :model-value="form.countryCode"
        aria-label="Quốc gia"
        placeholder="Tìm quốc gia…"
        :clearable="false"
        @update:model-value="onCountryChange"
      />
    </label>
    <label>
      Tỉnh / Thành phố
      <SearchableSelect
        :options="provinceOptions"
        :model-value="form.provinceCode"
        aria-label="Tỉnh / Thành phố"
        placeholder="Tìm tỉnh / thành phố…"
        @update:model-value="onProvinceChange"
      />
    </label>
    <label>
      Phường / Xã
      <SearchableSelect
        :options="wardOptions"
        :model-value="form.wardCode"
        aria-label="Phường / Xã"
        placeholder="Tìm phường / xã…"
        :disabled="!form.provinceCode"
        @update:model-value="onWardChange"
      />
    </label>

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
}

.form h3 {
  margin: 0.35rem 0 0;
  color: #475569;
  font-size: 0.95rem;
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

.avatarRow {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.avatarPreview {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
  background: #f1f5f9;
  overflow: hidden;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #64748b;
  flex-shrink: 0;
}

.avatarPreview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatarActions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.fileBtn {
  display: inline-flex;
  align-items: center;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font-weight: 500;
  cursor: pointer;
  background: #fff;
}

.fileBtn input {
  display: none;
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
</style>
