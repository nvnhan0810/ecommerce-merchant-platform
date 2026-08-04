<script setup lang="ts">
import { computed, ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute, useRouter } from 'vue-router'
import {
  GetMerchantUseCase,
  UpdateMerchantUseCase,
  UploadMerchantAvatarUseCase,
  DeleteMerchantAvatarUseCase,
} from '../application/merchant-use-cases'
import { HttpMerchantRepository } from '../infrastructure/http-merchant-repository'
import MerchantForm, { type MerchantFormPayload } from './MerchantForm.vue'

const route = useRoute()
const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpMerchantRepository()
const getMerchant = new GetMerchantUseCase(repo)
const updateMerchant = new UpdateMerchantUseCase(repo)
const uploadAvatar = new UploadMerchantAvatarUseCase(repo)
const deleteAvatar = new DeleteMerchantAvatarUseCase(repo)
const formError = ref('')

const merchantId = computed(() => String(route.params.id))

const { data: merchant, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'merchants', merchantId.value]),
  queryFn: () => getMerchant.execute(merchantId.value),
})

const saveMutation = useMutation({
  mutationFn: async (payload: MerchantFormPayload) => {
    let saved = await updateMerchant.execute({
      id: merchantId.value,
      email: payload.email,
      displayName: payload.displayName,
      password: payload.password.trim() || undefined,
      addressLine: payload.addressLine,
      countryCode: payload.countryCode,
      provinceCode: payload.provinceCode,
      wardCode: payload.wardCode,
    })
    if (payload.file) {
      saved = await uploadAvatar.execute(saved.id, payload.file)
    } else if (payload.removeAvatar) {
      saved = await deleteAvatar.execute(saved.id)
    }
    return saved
  },
  onSuccess: async (saved) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
    await router.push(`/merchants/${saved.id}`)
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
        <RouterLink class="back" :to="`/merchants/${merchantId}`">← Chi tiết</RouterLink>
        <h1>Sửa merchant</h1>
      </div>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <MerchantForm
      v-else-if="merchant"
      :initial="merchant"
      submit-label="Cập nhật"
      :pending="saveMutation.isPending.value"
      :error="formError"
      @submit="
        (p) => {
          formError = ''
          saveMutation.mutate(p)
        }
      "
      @cancel="router.push(`/merchants/${merchantId}`)"
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
