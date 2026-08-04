<script setup lang="ts">
import { ref } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import {
  CreateMerchantUseCase,
  UploadMerchantAvatarUseCase,
} from '../application/merchant-use-cases'
import { HttpMerchantRepository } from '../infrastructure/http-merchant-repository'
import MerchantForm, { type MerchantFormPayload } from './MerchantForm.vue'

const router = useRouter()
const queryClient = useQueryClient()
const repo = new HttpMerchantRepository()
const createMerchant = new CreateMerchantUseCase(repo)
const uploadAvatar = new UploadMerchantAvatarUseCase(repo)
const formError = ref('')

const saveMutation = useMutation({
  mutationFn: async (payload: MerchantFormPayload) => {
    let saved = await createMerchant.execute({
      email: payload.email,
      displayName: payload.displayName,
      password: payload.password,
      addressLine: payload.addressLine,
      countryCode: payload.countryCode,
      provinceCode: payload.provinceCode,
      wardCode: payload.wardCode,
    })
    if (payload.file) {
      saved = await uploadAvatar.execute(saved.id, payload.file)
    }
    return saved
  },
  onSuccess: async (merchant) => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'merchants'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
    await router.push(`/merchants/${merchant.id}`)
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
        <RouterLink class="back" to="/merchants">← Merchants</RouterLink>
        <h1>Tạo merchant</h1>
      </div>
    </header>
    <MerchantForm
      submit-label="Tạo merchant"
      :pending="saveMutation.isPending.value"
      :error="formError"
      @submit="
        (p) => {
          formError = ''
          saveMutation.mutate(p)
        }
      "
      @cancel="router.push('/merchants')"
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
