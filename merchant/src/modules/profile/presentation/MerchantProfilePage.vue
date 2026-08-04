<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  DeleteMerchantAvatarUseCase,
  GetMerchantProfileUseCase,
  UpdateMerchantProfileUseCase,
  UploadMerchantAvatarUseCase,
} from '../application/profile-use-cases'
import { HttpMerchantProfileRepository } from '../infrastructure/http-profile-repository'
import { fetchCountries, fetchProvinces, fetchWards } from '@/shared/geo'
import SearchableSelect from '@/shared/SearchableSelect.vue'
import { LocalStorageSessionStore } from '@/modules/auth/infrastructure/local-storage-session-store'
import { AuthSession } from '@/modules/auth/domain/session'

const repo = new HttpMerchantProfileRepository()
const getProfile = new GetMerchantProfileUseCase(repo)
const updateProfile = new UpdateMerchantProfileUseCase(repo)
const uploadAvatar = new UploadMerchantAvatarUseCase(repo)
const deleteAvatar = new DeleteMerchantAvatarUseCase(repo)
const sessionStore = new LocalStorageSessionStore()
const queryClient = useQueryClient()

const formError = ref('')
const successMessage = ref('')
const avatarFile = ref<File | null>(null)
const avatarPreview = ref('')
const removeAvatar = ref(false)

const form = reactive({
  displayName: '',
  password: '',
  addressLine: '',
  countryCode: 'VN',
  provinceCode: '',
  wardCode: '',
})

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['merchant', 'profile'],
  queryFn: () => getProfile.execute(),
})

watch(
  data,
  (profile) => {
    if (!profile) return
    form.displayName = profile.displayName
    form.password = ''
    form.addressLine = profile.addressLine
    form.countryCode = profile.countryCode || 'VN'
    form.provinceCode = profile.provinceCode
    form.wardCode = profile.wardCode
    avatarPreview.value = profile.avatarUrl
    avatarFile.value = null
    removeAvatar.value = false
  },
  { immediate: true },
)

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
  avatarFile.value = file
  removeAvatar.value = false
  if (file) avatarPreview.value = URL.createObjectURL(file)
}

function clearAvatar(): void {
  avatarFile.value = null
  avatarPreview.value = ''
  removeAvatar.value = true
}

const saveMutation = useMutation({
  mutationFn: async () => {
    let saved = await updateProfile.execute({
      displayName: form.displayName.trim(),
      password: form.password.trim() || undefined,
      addressLine: form.addressLine.trim(),
      countryCode: form.countryCode,
      provinceCode: form.provinceCode,
      wardCode: form.wardCode,
    })
    if (avatarFile.value) {
      saved = await uploadAvatar.execute(avatarFile.value)
    } else if (removeAvatar.value) {
      saved = await deleteAvatar.execute()
    }
    return saved
  },
  onSuccess: async (saved) => {
    form.password = ''
    avatarFile.value = null
    removeAvatar.value = false
    avatarPreview.value = saved.avatarUrl
    successMessage.value = 'Đã cập nhật thông tin gian hàng.'
    formError.value = ''
    const current = sessionStore.load()
    if (current) {
      sessionStore.save(
        new AuthSession(current.accessToken, saved.id, saved.email, saved.displayName, current.role),
      )
    }
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'profile'] })
  },
  onError: (e: Error) => {
    formError.value = e.message
    successMessage.value = ''
  },
})

async function onSubmit(): Promise<void> {
  formError.value = ''
  successMessage.value = ''
  await saveMutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header>
      <h1>Thông tin gian hàng</h1>
      <p>Cập nhật tên, địa chỉ và avatar hiển thị trên cửa hàng.</p>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>

    <form v-else class="form" @submit.prevent="onSubmit">
      <div class="avatarRow">
        <div class="avatarPreview" aria-hidden="true">
          <img v-if="avatarPreview" :src="avatarPreview" alt="" />
          <span v-else>{{ form.displayName?.charAt(0) || '?' }}</span>
        </div>
        <div class="avatarActions">
          <label class="fileBtn">
            Chọn avatar
            <input type="file" accept="image/jpeg,image/png,image/webp,image/gif" @change="onAvatarChange" />
          </label>
          <button v-if="avatarPreview" type="button" class="ghost" @click="clearAvatar">Xóa avatar</button>
        </div>
      </div>

      <label>
        Email
        <input :value="data?.email ?? ''" type="email" disabled />
      </label>
      <label>
        Tên gian hàng
        <input v-model="form.displayName" required />
      </label>
      <label>
        Mật khẩu mới
        <input
          v-model="form.password"
          type="password"
          placeholder="Để trống nếu giữ mật khẩu cũ"
          autocomplete="new-password"
        />
      </label>

      <h2>Địa chỉ</h2>
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

      <p v-if="formError" class="error" role="alert">{{ formError }}</p>
      <p v-if="successMessage" class="ok" role="status">{{ successMessage }}</p>
      <button type="submit" class="primary" :disabled="saveMutation.isPending.value">
        {{ saveMutation.isPending.value ? 'Đang lưu…' : 'Lưu thay đổi' }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 640px;
}

header h1 {
  margin: 0;
}

header p {
  margin: 0.35rem 0 0;
  color: #64748b;
}

.form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form h2 {
  margin: 0.35rem 0 0;
  font-size: 1rem;
  color: #475569;
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

input:disabled {
  background: #f8fafc;
  color: #64748b;
}

.avatarRow {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.avatarPreview {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 1px solid #e2e8f0;
  background: #f1f5f9;
  overflow: hidden;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #64748b;
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

button {
  border-radius: 8px;
  padding: 0.5rem 0.9rem;
  font: inherit;
  cursor: pointer;
  width: fit-content;
}

.primary {
  border: 0;
  background: #0c4a6e;
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

.ok {
  color: #166534;
  margin: 0;
}
</style>
