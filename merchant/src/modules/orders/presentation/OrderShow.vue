<script setup lang="ts">
import { computed, ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRoute } from 'vue-router'
import { GetOrderUseCase, UpdateOrderStatusUseCase } from '../application/order-use-cases'
import { HttpOrderRepository } from '../infrastructure/http-order-repository'
import OrderStatusBadge from './OrderStatusBadge.vue'

const route = useRoute()
const queryClient = useQueryClient()
const repo = new HttpOrderRepository()
const getOrder = new GetOrderUseCase(repo)
const updateStatus = new UpdateOrderStatusUseCase(repo)

const orderId = computed(() => String(route.params.id))
const statusMessage = ref('')
const cancelReason = ref('')
const showCancelForm = ref(false)

const { data: order, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['merchant', 'orders', orderId.value]),
  queryFn: () => getOrder.execute(orderId.value),
})

const history = computed(() => order.value?.history ?? [])
const deliveryEvents = computed(() => order.value?.deliveryEvents ?? [])
const canConfirm = computed(() => order.value?.status === 'new')
const canCancel = computed(() => order.value?.status === 'new')

function formatMoney(cents: number, currency: string): string {
  return `${cents.toLocaleString('vi-VN')} ${currency}`
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('vi-VN')
  } catch {
    return iso
  }
}

function actorLabel(name: string, email: string, role: string): string {
  const who = name || email || 'Không rõ'
  if (!role) return who
  return `${who} · ${role}`
}

const confirmMutation = useMutation({
  mutationFn: () => updateStatus.execute(orderId.value, 'confirmed'),
  onSuccess: async (order) => {
    statusMessage.value = order.deliveryTrackingCode
      ? `Đã xác nhận và tạo vận đơn ${order.deliveryTrackingCode}.`
      : 'Đã xác nhận và tạo vận đơn.'
    showCancelForm.value = false
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'orders'] })
  },
  onError: (err: Error) => {
    statusMessage.value = err.message
  },
})

const cancelMutation = useMutation({
  mutationFn: (reason: string) => updateStatus.execute(orderId.value, 'cancelled', reason),
  onSuccess: async () => {
    statusMessage.value = 'Đã huỷ đơn hàng.'
    cancelReason.value = ''
    showCancelForm.value = false
    await queryClient.invalidateQueries({ queryKey: ['merchant', 'orders'] })
  },
  onError: (err: Error) => {
    statusMessage.value = err.message
  },
})

const busy = computed(() => confirmMutation.isPending.value || cancelMutation.isPending.value)

async function onConfirm(): Promise<void> {
  statusMessage.value = ''
  await confirmMutation.mutateAsync()
}

async function onCancel(): Promise<void> {
  statusMessage.value = ''
  const reason = cancelReason.value.trim()
  if (!reason) {
    statusMessage.value = 'Vui lòng nhập lý do huỷ.'
    return
  }
  await cancelMutation.mutateAsync(reason)
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/orders">← Đơn hàng</RouterLink>
        <h1>{{ order?.code || 'Chi tiết đơn hàng' }}</h1>
        <div v-if="order" class="header-badge">
          <OrderStatusBadge :status="order.status" :label="order.statusLabel" />
        </div>
      </div>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <article v-else-if="order" class="detail">
      <dl class="meta">
        <div>
          <dt>Mã đơn</dt>
          <dd class="code">{{ order.code }}</dd>
        </div>
        <div>
          <dt>Mã vận đơn</dt>
          <dd class="code">{{ order.deliveryTrackingCode || '—' }}</dd>
        </div>
        <div>
          <dt>Đơn vị vận chuyển</dt>
          <dd>{{ order.deliveryCarrier || 'internal' }}</dd>
        </div>
        <div>
          <dt>Trạng thái</dt>
          <dd><OrderStatusBadge :status="order.status" :label="order.statusLabel" /></dd>
        </div>
        <div>
          <dt>Tổng tiền</dt>
          <dd>{{ formatMoney(order.totalCents, order.currency) }}</dd>
        </div>
        <div>
          <dt>Ghi chú</dt>
          <dd>{{ order.note || '—' }}</dd>
        </div>
        <div>
          <dt>Tạo lúc</dt>
          <dd>{{ formatDate(order.createdAt) }}</dd>
        </div>
        <div>
          <dt>Cập nhật</dt>
          <dd>{{ formatDate(order.updatedAt) }}</dd>
        </div>
      </dl>

      <div v-if="canConfirm || canCancel" class="status-form">
        <button
          v-if="canConfirm"
          type="button"
          class="primary"
          :disabled="busy"
          @click="onConfirm"
        >
          Xác nhận đơn
        </button>
        <button
          v-if="canCancel && !showCancelForm"
          type="button"
          class="danger"
          :disabled="busy"
          @click="showCancelForm = true"
        >
          Huỷ đơn
        </button>
        <div v-if="canCancel && showCancelForm" class="cancel-box">
          <label>
            Lý do huỷ (bắt buộc)
            <textarea v-model="cancelReason" rows="3" placeholder="VD: Hết hàng kho" />
          </label>
          <div class="cancel-actions">
            <button type="button" class="danger" :disabled="busy" @click="onCancel">
              Xác nhận huỷ
            </button>
            <button type="button" class="ghost" :disabled="busy" @click="showCancelForm = false">
              Đóng
            </button>
          </div>
        </div>
        <p v-if="statusMessage" class="hint">{{ statusMessage }}</p>
      </div>
      <p v-else-if="statusMessage" class="hint">{{ statusMessage }}</p>

      <div class="items">
        <h2>Sản phẩm</h2>
        <table>
          <thead>
            <tr>
              <th>Tên</th>
              <th>Đơn giá</th>
              <th>SL</th>
              <th>Thành tiền</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in order.items" :key="item.id">
              <td>{{ item.productName }}</td>
              <td>{{ formatMoney(item.unitPriceCents, order.currency) }}</td>
              <td>{{ item.quantity }}</td>
              <td>{{ formatMoney(item.lineTotalCents, order.currency) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="history">
        <h2>Timeline TMS</h2>
        <ol v-if="deliveryEvents.length" class="timeline">
          <li v-for="ev in deliveryEvents" :key="ev.id" class="event--delivery">
            <div class="dot" aria-hidden="true" />
            <div class="event-body">
              <div class="event-top">
                <strong>{{ ev.statusLabel || ev.statusCode }}</strong>
                <time>{{ formatDate(ev.occurredAt) }}</time>
              </div>
              <p class="event-msg">{{ ev.message }}</p>
              <p v-if="ev.reason" class="event-actor">Lý do: {{ ev.reason }}</p>
              <p class="event-actor">
                {{ ev.deliveryTrackingCode || '—' }} · {{ ev.source || 'tms' }}
              </p>
            </div>
          </li>
        </ol>
        <p v-else class="empty-history">Chưa có sự kiện vận chuyển.</p>
      </div>

      <div class="history">
        <h2>Lịch sử cập nhật</h2>
        <ol v-if="history.length" class="timeline">
          <li v-for="ev in history" :key="ev.id" :class="`event--${ev.eventType}`">
            <div class="dot" aria-hidden="true" />
            <div class="event-body">
              <div class="event-top">
                <strong>{{ ev.eventLabel }}</strong>
                <time>{{ formatDate(ev.createdAt) }}</time>
              </div>
              <p class="event-msg">{{ ev.message }}</p>
              <div v-if="ev.toStatus" class="event-status">
                <OrderStatusBadge
                  v-if="ev.fromStatus"
                  :status="ev.fromStatus"
                  :label="ev.fromStatusLabel"
                />
                <span v-if="ev.fromStatus" class="arrow">→</span>
                <OrderStatusBadge :status="ev.toStatus" :label="ev.toStatusLabel" />
              </div>
              <p class="event-actor">
                {{ actorLabel(ev.actorName, ev.actorEmail, ev.actorRole) }}
              </p>
            </div>
          </li>
        </ol>
        <p v-else class="empty-history">Chưa có lịch sử cập nhật.</p>
      </div>
    </article>
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
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.04em;
}

.header-badge {
  margin-top: 0.55rem;
}

.back {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9rem;
}

.back:hover {
  color: #0f172a;
}

.detail {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.meta {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0.85rem;
}

dt {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #64748b;
  margin-bottom: 0.2rem;
}

dd {
  margin: 0;
  font-size: 1rem;
  color: #0f172a;
}

.code {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  letter-spacing: 0.04em;
}

.status-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-start;
  border-top: 1px solid #e2e8f0;
  padding-top: 1rem;
}

.cancel-box {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.cancel-box label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.cancel-box textarea {
  font: inherit;
  text-transform: none;
  letter-spacing: normal;
  color: #0f172a;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.55rem 0.65rem;
  resize: vertical;
  min-height: 4.5rem;
}

.cancel-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.hint {
  margin: 0;
  color: #64748b;
  font-size: 0.9rem;
  width: 100%;
}

.items h2,
.history h2 {
  margin: 0 0 0.65rem;
  font-size: 1rem;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}

th,
td {
  text-align: left;
  padding: 0.55rem 0.4rem;
  border-bottom: 1px solid #e2e8f0;
}

th {
  color: #64748b;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.history {
  border-top: 1px solid #e2e8f0;
  padding-top: 1rem;
}

.timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  position: relative;
}

.timeline li {
  display: grid;
  grid-template-columns: 1.1rem 1fr;
  gap: 0.75rem;
  position: relative;
  padding-bottom: 1rem;
}

.timeline li:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 0.4rem;
  top: 1rem;
  bottom: 0;
  width: 2px;
  background: #e2e8f0;
}

.dot {
  width: 0.85rem;
  height: 0.85rem;
  border-radius: 999px;
  margin-top: 0.25rem;
  background: #94a3b8;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px #cbd5e1;
  z-index: 1;
}

.event--created .dot {
  background: #2563eb;
  box-shadow: 0 0 0 1px #93c5fd;
}

.event--status_changed .dot {
  background: #7c3aed;
  box-shadow: 0 0 0 1px #c4b5fd;
}

.event--cancelled .dot {
  background: #64748b;
  box-shadow: 0 0 0 1px #cbd5e1;
}

.event--delivery .dot {
  background: #ea580c;
  box-shadow: 0 0 0 1px #fdba74;
}

.event-body {
  min-width: 0;
}

.event-top {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: baseline;
}

.event-top time {
  color: #94a3b8;
  font-size: 0.85rem;
}

.event-msg {
  margin: 0.25rem 0 0;
  color: #334155;
}

.event-status {
  margin-top: 0.45rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.arrow {
  color: #94a3b8;
  font-size: 0.85rem;
}

.event-actor {
  margin: 0.4rem 0 0;
  color: #64748b;
  font-size: 0.88rem;
}

.empty-history {
  margin: 0;
  color: #94a3b8;
}

.error {
  color: #b91c1c;
}

button,
a.ghost {
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
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

.danger {
  border: 0;
  background: #b91c1c;
  color: #fff;
}

.ghost {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #334155;
}
</style>
