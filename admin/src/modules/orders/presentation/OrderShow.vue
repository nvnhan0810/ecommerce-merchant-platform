<script setup lang="ts">
import { computed, ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRoute } from 'vue-router'
import { GetOrderUseCase } from '../application/order-use-cases'
import { HttpOrderRepository } from '../infrastructure/http-order-repository'
import OrderStatusBadge from './OrderStatusBadge.vue'
import { ListUsersUseCase } from '@/modules/users/application/user-use-cases'
import { HttpUserRepository } from '@/modules/users/infrastructure/http-user-repository'
import { ListMerchantsUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'

const route = useRoute()
const repo = new HttpOrderRepository()
const getOrder = new GetOrderUseCase(repo)
const listUsers = new ListUsersUseCase(new HttpUserRepository())
const listMerchants = new ListMerchantsUseCase(new HttpMerchantRepository())

const orderId = computed(() => String(route.params.id))
const activeTab = ref<'details' | 'tracking' | 'history'>('details')

const { data: order, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'orders', orderId.value]),
  queryFn: () => getOrder.execute(orderId.value),
})

const { data: users } = useQuery({
  queryKey: ['admin', 'users'],
  queryFn: () => listUsers.execute(),
})

const { data: merchants } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => listMerchants.execute(),
})

const userLabel = computed(() => {
  const id = order.value?.userId
  if (!id) return ''
  const u = (users.value ?? []).find((x) => x.id === id)
  return u ? `${u.displayName} (${u.email})` : id
})

const merchantLabel = computed(() => {
  const id = order.value?.merchantId
  if (!id) return ''
  const m = (merchants.value ?? []).find((x) => x.id === id)
  return m ? `${m.displayName} (${m.email})` : id
})

const history = computed(() => order.value?.history ?? [])
const deliveryEvents = computed(() => order.value?.deliveryEvents ?? [])

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
</script>

<template>
  <section class="panel">
    <header class="header">
      <div>
        <RouterLink class="back" to="/orders">← Orders</RouterLink>
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
          <dt>Tổng tiền</dt>
          <dd class="total-price">{{ formatMoney(order.totalCents, order.currency) }}</dd>
        </div>
        <div class="meta-item">
          <dt>User</dt>
          <dd>{{ userLabel }}</dd>
        </div>
        <div class="meta-item">
          <dt>Merchant</dt>
          <dd>{{ merchantLabel }}</dd>
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
        <h3>Thông tin giao hàng</h3>
        <dl class="shipping-dl">
          <div class="shipping-row">
            <div>
              <dt>Người nhận</dt>
              <dd>{{ order.shippingName || '—' }}</dd>
            </div>
            <div>
              <dt>Số điện thoại</dt>
              <dd>{{ order.shippingPhone || '—' }}</dd>
            </div>
          </div>
          <div>
            <dt>Địa chỉ</dt>
            <dd>{{ order.shippingAddress || '—' }}</dd>
          </div>
        </dl>
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
        <section class="carrier-section">
          <h2>Thông tin vận chuyển</h2>
          <dl class="shipping-dl">
            <div>
              <dt>Mã vận đơn</dt>
              <dd class="code">{{ order.deliveryTrackingCode || '—' }}</dd>
            </div>
            <div>
              <dt>Đơn vị vận chuyển</dt>
              <dd>{{ order.deliveryCarrier || 'internal' }}</dd>
            </div>
          </dl>
        </section>
        <div class="history">
          <h2>Lịch trình</h2>
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
          <p v-else class="empty-history">Chưa có sự kiện vận chuyển.</p>
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
  gap: 1.25rem;
  padding: 0 0 1.25rem;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 1.25rem;
}

.address-box {
  padding: 0 0 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 1.5rem;
}

.address-box h3 {
  margin: 0 0 0.75rem 0;
  font-size: 0.95rem;
  color: #0f172a;
}

.carrier-section {
  margin-bottom: 1.5rem;
}

.carrier-section h2,
.history h2 {
  margin: 0 0 0.75rem;
  font-size: 1.05rem;
}

.shipping-dl {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.shipping-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.shipping-dl > div:not(.shipping-row),
.shipping-row > div {
  display: grid;
  grid-template-columns: 8.5rem 1fr;
  gap: 0.5rem;
}

.shipping-dl dt {
  color: #64748b;
  font-weight: 500;
}

.shipping-dl dd {
  margin: 0;
  color: #0f172a;
}

@media (max-width: 640px) {
  .shipping-row {
    grid-template-columns: 1fr;
  }

  .shipping-dl > div:not(.shipping-row),
  .shipping-row > div {
    grid-template-columns: 1fr;
    gap: 0.15rem;
  }
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

.actions-row {
  display: flex;
  justify-content: flex-end;
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
  padding: 1.5rem 0;
  text-align: left;
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
