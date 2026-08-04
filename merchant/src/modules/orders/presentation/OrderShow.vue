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
const activeTab = ref<'details' | 'tracking' | 'history'>('details')
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
      <div class="meta-grid">
        <div class="meta-item">
          <dt>Mã vận đơn</dt>
          <dd class="code">{{ order.deliveryTrackingCode || '—' }}</dd>
        </div>
        <div class="meta-item">
          <dt>Đơn vị vận chuyển</dt>
          <dd>{{ order.deliveryCarrier || 'internal' }}</dd>
        </div>
        <div class="meta-item">
          <dt>Tổng tiền</dt>
          <dd class="total-price">{{ formatMoney(order.totalCents, order.currency) }}</dd>
        </div>
        <div class="meta-item">
          <dt>Tạo lúc</dt>
          <dd>{{ formatDate(order.createdAt) }}</dd>
        </div>
        <div class="meta-item">
          <dt>Cập nhật</dt>
          <dd>{{ formatDate(order.updatedAt) }}</dd>
        </div>
        <div class="meta-item">
          <dt>Ghi chú</dt>
          <dd>{{ order.note || '—' }}</dd>
        </div>
      </div>

      <div class="address-box">
        <h3>Địa chỉ nhận hàng</h3>
        <p><strong>{{ order.shippingName || '—' }}</strong> - {{ order.shippingPhone || '—' }}</p>
        <p>{{ order.shippingAddress || '—' }}</p>
      </div>

      <div class="tabs">
        <button
          type="button"
          class="tab-btn"
          :class="{ 'active-tab': activeTab === 'details' }"
          @click="activeTab = 'details'"
        >
          Chi tiết đơn hàng
        </button>
        <button
          type="button"
          class="tab-btn"
          :class="{ 'active-tab': activeTab === 'tracking' }"
          @click="activeTab = 'tracking'"
        >
          Theo dõi vận chuyển
        </button>
        <button
          type="button"
          class="tab-btn"
          :class="{ 'active-tab': activeTab === 'history' }"
          @click="activeTab = 'history'"
        >
          Lịch sử cập nhật
        </button>
      </div>

      <div v-if="activeTab === 'details'" class="tab-content">
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
          <div class="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Tên sản phẩm</th>
                  <th>Đơn giá</th>
                  <th class="text-center">Số lượng</th>
                  <th class="text-right">Thành tiền</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in order.items" :key="item.id">
                  <td>{{ item.productName }}</td>
                  <td>{{ formatMoney(item.unitPriceCents, order.currency) }}</td>
                  <td class="text-center">{{ item.quantity }}</td>
                  <td class="text-right">{{ formatMoney(item.lineTotalCents, order.currency) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'tracking'" class="tab-content">
        <div class="history">
          <ol v-if="deliveryEvents.length" class="timeline">
            <li v-for="ev in deliveryEvents" :key="ev.id" class="event--delivery">
              <div class="dot" aria-hidden="true" />
              <div class="event-body">
                <div class="event-top">
                  <strong>{{ ev.statusLabel || ev.statusCode }}</strong>
                  <time>{{ formatDate(ev.occurredAt) }}</time>
                </div>
                <p class="event-msg">{{ ev.message }}</p>
                <p v-if="ev.reason" class="event-actor reason-box">Lý do: {{ ev.reason }}</p>
                <p class="event-actor">
                  {{ ev.deliveryTrackingCode || '—' }} · {{ ev.source || 'tms' }}
                </p>
              </div>
            </li>
          </ol>
          <div v-else class="empty-state">
            <p class="empty-history">Chưa có sự kiện vận chuyển.</p>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'history'" class="tab-content">
        <div class="history">
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
          <div v-else class="empty-state">
            <p class="empty-history">Chưa có lịch sử cập nhật.</p>
          </div>
        </div>
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
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1.5rem;
  background: #f8fafc;
  padding: 1.25rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  margin-bottom: 1rem;
}

.address-box {
  background: #f8fafc;
  padding: 1.25rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  margin-bottom: 2rem;
}

.address-box h3 {
  margin: 0 0 0.5rem 0;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: #64748b;
}

.address-box p {
  margin: 0.25rem 0;
  color: #0f172a;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 1.5rem;
}

.tab-btn {
  background: none;
  border: none;
  padding: 0.75rem 1.25rem;
  font: inherit;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: all 0.2s;
  border-radius: 0;
}

.tab-btn:hover {
  color: #fff;
  background-color: #0f172a;
  border-bottom-color: #0f172a;
}

.active-tab {
  color: #0f172a;
  border-bottom-color: #0f172a;
}

.active-tab:hover {
  color: #fff;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}

dt {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: #64748b;
  margin-bottom: 0;
}

dd {
  margin: 0;
  font-size: 1rem;
  color: #0f172a;
}

.total-price {
  color: #0f766e;
  font-size: 1.1rem;
  font-weight: 600;
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
  background: #f8fafc;
  padding: 1.25rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
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

.table-wrapper {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
  min-width: 500px;
}

th,
td {
  text-align: left;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #e2e8f0;
}

th {
  color: #64748b;
  font-weight: 600;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: #f8fafc;
}

tr:last-child td {
  border-bottom: none;
}

.text-center {
  text-align: center;
}

.text-right {
  text-align: right;
}

.history {
  display: flex;
  flex-direction: column;
}

.timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  position: relative;
}

.timeline li {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding-left: 1.25rem;
  border-left: 2px solid #cbd5e1;
  position: relative;
}

.timeline li::before {
  content: '';
  position: absolute;
  left: -5px;
  top: 0.25rem;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
}

.event--created::before { background: #2563eb !important; }
.event--status_changed::before { background: #7c3aed !important; }
.event--cancelled::before { background: #64748b !important; }
.event--delivery::before { background: #ea580c !important; }

.dot {
  display: none;
}

.event-body {
  min-width: 0;
}

.event-top {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.event-top time {
  color: #94a3b8;
  font-size: 0.9rem;
}

.event-msg {
  margin: 0;
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

.reason-box {
  color: #b91c1c;
  background: #fef2f2;
  padding: 0.35rem 0.5rem;
  border-radius: 6px;
  display: inline-block;
  align-self: flex-start;
}

.empty-state {
  padding: 3rem;
  text-align: center;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px dashed #e2e8f0;
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

.items h2 {
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 1.1rem;
  color: #0f172a;
}
</style>
