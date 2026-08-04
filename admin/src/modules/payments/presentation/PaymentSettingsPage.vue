<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { GetPaymentSettingsUseCase, UpdatePaymentSettingsUseCase } from '../application/payment-settings-use-cases'
import { HttpPaymentSettingsRepository } from '../infrastructure/http-payment-settings-repository'
import type { OnePayDemoCredentials, OnePayGatewayInput } from '../domain/payment-settings'
import PaymentMethodSection from './PaymentMethodSection.vue'

const repo = new HttpPaymentSettingsRepository()
const getSettings = new GetPaymentSettingsUseCase(repo)
const updateSettings = new UpdatePaymentSettingsUseCase(repo)
const queryClient = useQueryClient()

type GatewayForm = OnePayGatewayInput & { secretDirty: boolean }

const shared = reactive({
  onepayReturnUrl: '',
  onepayIpnUrl: '',
})

const domestic = reactive<GatewayForm>({
  enabled: false,
  merchantId: '',
  accessCode: '',
  hashSecret: '',
  paymentUrl: 'https://mtf.onepay.vn/onecomm-pay/vpc.op',
  secretDirty: false,
})

const international = reactive<GatewayForm>({
  enabled: false,
  merchantId: '',
  accessCode: '',
  hashSecret: '',
  paymentUrl: 'https://mtf.onepay.vn/vpcpay/vpcpay.op',
  secretDirty: false,
})

const message = ref('')

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'payment-settings'],
  queryFn: () => getSettings.execute(),
})

function applyGateway(target: GatewayForm, source: OnePayGatewayInput): void {
  target.enabled = source.enabled
  target.merchantId = source.merchantId
  target.accessCode = source.accessCode
  target.hashSecret = ''
  target.paymentUrl = source.paymentUrl
  target.secretDirty = false
}

watch(
  data,
  (value) => {
    if (!value) return
    shared.onepayReturnUrl = value.onepayReturnUrl
    shared.onepayIpnUrl = value.onepayIpnUrl
    applyGateway(domestic, value.onepayDomestic)
    applyGateway(international, value.onepayInternational)
  },
  { immediate: true },
)

function badgeFor(ready: boolean, enabled: boolean): { text: string; tone: 'ready' | 'pending' | 'muted' } {
  if (ready) return { text: 'Sẵn sàng', tone: 'ready' }
  if (enabled) return { text: 'Thiếu cấu hình', tone: 'pending' }
  return { text: 'Đang tắt', tone: 'muted' }
}

const domesticBadge = computed(() =>
  data.value ? badgeFor(data.value.onepayDomestic.ready, data.value.onepayDomestic.enabled) : badgeFor(false, false),
)
const internationalBadge = computed(() =>
  data.value
    ? badgeFor(data.value.onepayInternational.ready, data.value.onepayInternational.enabled)
    : badgeFor(false, false),
)

const mutation = useMutation({
  mutationFn: () =>
    updateSettings.execute({
      onepayReturnUrl: shared.onepayReturnUrl.trim(),
      onepayIpnUrl: shared.onepayIpnUrl.trim(),
      onepayDomestic: {
        enabled: domestic.enabled,
        merchantId: domestic.merchantId.trim(),
        accessCode: domestic.accessCode.trim(),
        hashSecret: domestic.secretDirty ? domestic.hashSecret.trim() : '',
        paymentUrl: domestic.paymentUrl.trim(),
      },
      onepayInternational: {
        enabled: international.enabled,
        merchantId: international.merchantId.trim(),
        accessCode: international.accessCode.trim(),
        hashSecret: international.secretDirty ? international.hashSecret.trim() : '',
        paymentUrl: international.paymentUrl.trim(),
      },
    }),
  onSuccess: async () => {
    message.value = 'Đã lưu cấu hình thanh toán.'
    await queryClient.invalidateQueries({ queryKey: ['admin', 'payment-settings'] })
  },
  onError: (err) => {
    message.value = (err as Error).message
  },
})

function fillDemo(target: GatewayForm, demo: OnePayDemoCredentials | undefined): void {
  if (!demo) return
  target.merchantId = demo.merchantId
  target.accessCode = demo.accessCode
  target.hashSecret = demo.hashSecret
  target.paymentUrl = demo.paymentUrl
  target.enabled = true
  target.secretDirty = true
  message.value = 'Đã điền cấu hình demo OnePay. Nhấn Lưu để áp dụng.'
}

function submit(): void {
  message.value = ''
  mutation.mutate()
}
</script>

<template>
  <section class="panel">
    <header>
      <h1>Thanh toán</h1>
      <p class="hint">
        OnePay tách thành 2 cổng: nội địa (ATM) và quốc tế (Visa/Master/JCB). Có thể điền cấu hình demo công khai của OnePay.
        · <RouterLink to="/payment-callbacks">Xem IPN / callbacks</RouterLink>
      </p>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>

    <form v-else class="providers" @submit.prevent="submit">
      <PaymentMethodSection
        id="cod"
        title="Thanh toán khi giao hàng (COD)"
        description="Khách trả tiền mặt khi nhận hàng. Không cần cấu hình cổng ngoài."
        badge="Luôn bật"
        badge-tone="ready"
        :default-open="false"
      >
        <p class="info">
          COD luôn khả dụng trên checkout. Khi đơn giao thành công, hệ thống đánh dấu đã thanh toán.
        </p>
      </PaymentMethodSection>

      <PaymentMethodSection
        id="onepay-shared"
        title="OnePay — URL dùng chung"
        description="Return URL và IPN dùng chung cho cả nội địa và quốc tế."
        badge="Chung"
        badge-tone="muted"
        :default-open="true"
      >
        <div class="form">
          <label>
            Return URL (API)
            <input v-model="shared.onepayReturnUrl" type="url" />
            <small>OnePay redirect về API; API chuyển tiếp về storefront.</small>
          </label>
          <label>
            IPN URL
            <input v-model="shared.onepayIpnUrl" type="url" />
            <small>URL thông báo server-to-server từ OnePay.</small>
          </label>
        </div>
      </PaymentMethodSection>

      <PaymentMethodSection
        id="onepay-domestic"
        title="OnePay nội địa"
        description="Thẻ ATM nội địa / Internet Banking. URL test: onecomm-pay/vpc.op"
        :badge="domesticBadge.text"
        :badge-tone="domesticBadge.tone"
        :default-open="true"
      >
        <div class="form">
          <div class="toolbar">
            <label class="switch">
              <input v-model="domestic.enabled" type="checkbox" />
              <span>Bật OnePay nội địa</span>
            </label>
            <button type="button" class="secondary" @click="fillDemo(domestic, data?.demoDomestic)">
              Điền demo OnePay
            </button>
          </div>

          <p class="demo-note">
            {{ data?.demoDomestic.note || 'Sandbox OnePay nội địa.' }}
            <br />
            Demo: Merchant <code>TESTONEPAY</code>, Access <code>6BEB2546</code>, URL
            <code>https://mtf.onepay.vn/onecomm-pay/vpc.op</code>
          </p>

          <label>
            Merchant ID (vpc_Merchant)
            <input v-model="domestic.merchantId" type="text" autocomplete="off" />
          </label>
          <label>
            Access Code (vpc_AccessCode)
            <input v-model="domestic.accessCode" type="text" autocomplete="off" />
          </label>
          <label>
            Hash Secret
            <input
              v-model="domestic.hashSecret"
              type="password"
              autocomplete="new-password"
              :placeholder="data?.onepayDomestic.hashSecret ? '******** (để trống nếu không đổi)' : 'Nhập hash secret'"
              @input="domestic.secretDirty = true"
            />
          </label>
          <label>
            Payment URL
            <input v-model="domestic.paymentUrl" type="url" />
          </label>
        </div>
      </PaymentMethodSection>

      <PaymentMethodSection
        id="onepay-international"
        title="OnePay quốc tế"
        description="Thẻ Visa / Mastercard / JCB. URL test: vpcpay/vpcpay.op"
        :badge="internationalBadge.text"
        :badge-tone="internationalBadge.tone"
        :default-open="true"
      >
        <div class="form">
          <div class="toolbar">
            <label class="switch">
              <input v-model="international.enabled" type="checkbox" />
              <span>Bật OnePay quốc tế</span>
            </label>
            <button type="button" class="secondary" @click="fillDemo(international, data?.demoInternational)">
              Điền demo OnePay
            </button>
          </div>

          <p class="demo-note">
            {{ data?.demoInternational.note || 'Sandbox OnePay quốc tế.' }}
            <br />
            Demo: Merchant <code>TESTONEPAY</code>, Access <code>6BEB2546</code>, URL
            <code>https://mtf.onepay.vn/vpcpay/vpcpay.op</code>
          </p>

          <label>
            Merchant ID (vpc_Merchant)
            <input v-model="international.merchantId" type="text" autocomplete="off" />
          </label>
          <label>
            Access Code (vpc_AccessCode)
            <input v-model="international.accessCode" type="text" autocomplete="off" />
          </label>
          <label>
            Hash Secret
            <input
              v-model="international.hashSecret"
              type="password"
              autocomplete="new-password"
              :placeholder="
                data?.onepayInternational.hashSecret ? '******** (để trống nếu không đổi)' : 'Nhập hash secret'
              "
              @input="international.secretDirty = true"
            />
          </label>
          <label>
            Payment URL
            <input v-model="international.paymentUrl" type="url" />
          </label>
        </div>
      </PaymentMethodSection>

      <p v-if="message" class="message">{{ message }}</p>

      <div class="actions">
        <button type="submit" :disabled="mutation.isPending.value">
          {{ mutation.isPending.value ? 'Đang lưu…' : 'Lưu cấu hình' }}
        </button>
      </div>
    </form>
  </section>
</template>

<style scoped>
.panel {
  background: transparent;
  padding: 0;
}

header {
  margin-bottom: 1rem;
}

header h1 {
  margin: 0;
  font-size: 1.4rem;
}

.hint {
  margin: 0.4rem 0 0;
  color: #64748b;
}

.providers {
  display: grid;
  gap: 0.85rem;
  max-width: 760px;
}

.form {
  display: grid;
  gap: 0.9rem;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  justify-content: space-between;
}

label {
  display: grid;
  gap: 0.35rem;
  font-weight: 600;
  font-size: 0.92rem;
}

input[type='text'],
input[type='url'],
input[type='password'] {
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 0.65rem 0.75rem;
  font: inherit;
}

.switch {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-weight: 600;
}

.switch input {
  width: 1.1rem;
  height: 1.1rem;
}

small {
  color: #64748b;
  font-weight: 400;
}

.info,
.demo-note {
  margin: 0;
  color: #475569;
  line-height: 1.5;
  font-size: 0.9rem;
}

.demo-note code {
  font-size: 0.82rem;
  background: #f1f5f9;
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
}

.actions {
  margin-top: 0.25rem;
}

button {
  border: 0;
  background: #0f766e;
  color: #fff;
  border-radius: 10px;
  padding: 0.65rem 1rem;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

button.secondary {
  background: #fff;
  color: #0f766e;
  border: 1px solid #99f6e4;
}

button:disabled {
  opacity: 0.7;
  cursor: wait;
}

.ghost {
  background: transparent;
  color: #0f766e;
  border: 1px solid #99f6e4;
}

.message {
  margin: 0;
  color: #0f766e;
}

.error {
  color: #b91c1c;
}
</style>
