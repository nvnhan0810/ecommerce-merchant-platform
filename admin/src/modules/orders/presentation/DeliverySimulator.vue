<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute } from 'vue-router'
import {
  ListOrdersUseCase,
  SimulateDeliveryUseCase,
} from '../application/order-use-cases'
import { HttpOrderRepository } from '../infrastructure/http-order-repository'
import {
  DELIVERY_STATUS_OPTIONS,
  type DeliveryStatusCode,
} from '../domain/order'
import OrderStatusBadge from './OrderStatusBadge.vue'

const route = useRoute()
const queryClient = useQueryClient()
const repo = new HttpOrderRepository()
const listOrders = new ListOrdersUseCase(repo)
const simulate = new SimulateDeliveryUseCase(repo)

const selectedOrderId = ref('')
const trackingCode = ref('')
const carrier = ref('internal')
const status = ref<DeliveryStatusCode>('accepted')
const message = ref('')
const reason = ref('')
const feedback = ref('')

const { data: orders, isLoading } = useQuery({
  queryKey: ['admin', 'orders', 'simulator'],
  queryFn: () => listOrders.execute(),
})

watch(
  () => route.query.orderId,
  (value) => {
    if (typeof value === 'string' && value) {
      selectedOrderId.value = value
    }
  },
  { immediate: true },
)

watch(
  [orders, selectedOrderId],
  () => {
    const order = (orders.value ?? []).find((o) => o.id === selectedOrderId.value)
    if (!order) return
    if (!trackingCode.value && order.deliveryTrackingCode) {
      trackingCode.value = order.deliveryTrackingCode
    }
    if (order.deliveryCarrier) {
      carrier.value = order.deliveryCarrier
    }
  },
  { immediate: true },
)

const selectedOrder = computed(() =>
  (orders.value ?? []).find((o) => o.id === selectedOrderId.value),
)

const mutation = useMutation({
  mutationFn: () =>
    simulate.execute(selectedOrderId.value, {
      deliveryTrackingCode: trackingCode.value.trim(),
      deliveryCarrier: carrier.value.trim() || 'internal',
      status: status.value,
      message: message.value.trim(),
      reason: reason.value.trim(),
      occurredAt: new Date().toISOString(),
    }),
  onSuccess: async (order) => {
    feedback.value = `Đã ghi sự kiện ${status.value}. Trạng thái đơn: ${order.statusLabel}.`
    await queryClient.invalidateQueries({ queryKey: ['admin', 'orders'] })
  },
  onError: (err: Error) => {
    feedback.value = err.message
  },
})

async function onSubmit(): Promise<void> {
  feedback.value = ''
  if (!selectedOrderId.value) {
    feedback.value = 'Chọn đơn hàng trước.'
    return
  }
  await mutation.mutateAsync()
}
</script>

<template>
  <section class="panel">
    <header>
      <h1>Delivery simulator</h1>
      <p class="sub">Gửi sự kiện TMS giả lập vào cùng pipeline webhook (`ApplyDeliveryEvent`).</p>
    </header>

    <p v-if="isLoading">Đang tải đơn hàng…</p>
    <form v-else class="form" @submit.prevent="onSubmit">
      <label>
        Đơn hàng
        <select v-model="selectedOrderId" required>
          <option value="" disabled>Chọn đơn…</option>
          <option v-for="o in orders ?? []" :key="o.id" :value="o.id">
            {{ o.code }} · {{ o.statusLabel }}
          </option>
        </select>
      </label>

      <div v-if="selectedOrder" class="current">
        <OrderStatusBadge :status="selectedOrder.status" :label="selectedOrder.statusLabel" />
        <span class="mono">{{ selectedOrder.deliveryTrackingCode || 'Chưa có mã vận đơn' }}</span>
        <div v-if="selectedOrder.shippingName" class="shipping-info">
          Giao đến: <strong>{{ selectedOrder.shippingName }}</strong> - {{ selectedOrder.shippingPhone }}
          <br />
          {{ selectedOrder.shippingAddress }}
        </div>
      </div>

      <label>
        Mã vận đơn (delivery_tracking_code)
        <input v-model="trackingCode" type="text" placeholder="GHN123456" />
      </label>

      <label>
        Carrier
        <select v-model="carrier">
          <option value="internal">internal</option>
          <option value="ghn">ghn</option>
          <option value="ghtk">ghtk</option>
        </select>
      </label>

      <label>
        Trạng thái TMS
        <select v-model="status">
          <option v-for="opt in DELIVERY_STATUS_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }} ({{ opt.value }})
          </option>
        </select>
      </label>

      <label>
        Message
        <input v-model="message" type="text" placeholder="Đã tiếp nhận" />
      </label>

      <label>
        Reason
        <input v-model="reason" type="text" placeholder="Tuỳ chọn" />
      </label>

      <button type="submit" class="primary" :disabled="mutation.isPending.value">
        Gửi sự kiện
      </button>
      <p v-if="feedback" class="hint">{{ feedback }}</p>
    </form>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 40rem;
}

header h1 {
  margin: 0;
}

.sub {
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
  gap: 0.9rem;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

input,
select {
  font: inherit;
  text-transform: none;
  letter-spacing: normal;
  color: #0f172a;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.5rem 0.65rem;
  background: #fff;
}

.current {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.mono {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.03em;
  color: #334155;
}

.shipping-info {
  width: 100%;
  font-size: 0.9rem;
  color: #475569;
  background: #f8fafc;
  padding: 0.5rem;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  margin-top: 0.25rem;
}

.primary {
  border: 0;
  background: #0f172a;
  color: #fff;
  border-radius: 8px;
  padding: 0.55rem 0.9rem;
  font: inherit;
  cursor: pointer;
  align-self: flex-start;
}

.primary:disabled {
  opacity: 0.65;
  cursor: wait;
}

.hint {
  margin: 0;
  color: #64748b;
}
</style>
