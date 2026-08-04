<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { GetPaymentCallbackUseCase } from '../application/payment-callback-use-cases'
import { HttpPaymentCallbackRepository } from '../infrastructure/http-payment-callback-repository'

const route = useRoute()
const id = computed(() => String(route.params.id ?? ''))
const getCallback = new GetPaymentCallbackUseCase(new HttpPaymentCallbackRepository())

const { data: item, isLoading, isError, error, refetch } = useQuery({
  queryKey: computed(() => ['admin', 'payment-callbacks', id.value]),
  queryFn: () => getCallback.execute(id.value),
  enabled: computed(() => Boolean(id.value)),
})

const prettyPayload = computed(() => {
  if (!item.value) return ''
  try {
    return JSON.stringify(item.value.rawPayload ?? {}, null, 2)
  } catch {
    return String(item.value.rawPayload)
  }
})

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('vi-VN')
  } catch {
    return iso
  }
}
</script>

<template>
  <section class="panel">
    <RouterLink class="back" to="/payment-callbacks">← Payment callbacks</RouterLink>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>

    <article v-else-if="item" class="card">
      <header class="header">
        <div>
          <h1>{{ item.providerLabel }} · {{ item.channelLabel }}</h1>
          <p class="muted">{{ formatDate(item.createdAt) }}</p>
        </div>
        <div class="pills">
          <span class="pill" :data-ok="item.success ? '1' : '0'">
            {{ item.success ? (item.paid ? 'Paid' : 'Failed txn') : 'Error' }}
          </span>
          <span v-if="item.httpMethod" class="pill">{{ item.httpMethod }}</span>
        </div>
      </header>

      <dl class="grid">
        <div>
          <dt>Merch txn ref</dt>
          <dd class="code">{{ item.merchTxnRef || '—' }}</dd>
        </div>
        <div>
          <dt>Response code</dt>
          <dd>{{ item.responseCode || '—' }}</dd>
        </div>
        <div>
          <dt>Payment method</dt>
          <dd>{{ item.paymentMethodLabel || item.paymentMethod || '—' }}</dd>
        </div>
        <div>
          <dt>Payment ID</dt>
          <dd class="code">{{ item.paymentId || '—' }}</dd>
        </div>
        <div class="full">
          <dt>Message</dt>
          <dd>{{ item.message || '—' }}</dd>
        </div>
        <div v-if="item.errorMessage" class="full">
          <dt>Error</dt>
          <dd class="err">{{ item.errorMessage }}</dd>
        </div>
      </dl>

      <div class="payload">
        <h2>Raw payload</h2>
        <pre>{{ prettyPayload }}</pre>
      </div>
    </article>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 960px;
}

.back {
  color: #64748b;
  text-decoration: none;
  width: fit-content;
}

.back:hover {
  color: #0f172a;
}

.card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.header h1 {
  margin: 0;
  font-size: 1.35rem;
}

.muted {
  margin: 0.35rem 0 0;
  color: #94a3b8;
}

.pills {
  display: flex;
  gap: 0.4rem;
  flex-wrap: wrap;
  align-items: flex-start;
}

.pill {
  display: inline-flex;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 650;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #475569;
}

.pill[data-ok='1'] {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

.pill[data-ok='0'] {
  background: #fef2f2;
  color: #b91c1c;
  border-color: #fecaca;
}

.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem 1.25rem;
  margin: 0;
}

.grid .full {
  grid-column: 1 / -1;
}

dt {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: #94a3b8;
}

dd {
  margin: 0.25rem 0 0;
  color: #0f172a;
}

.code {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  word-break: break-all;
}

.err {
  color: #b91c1c;
}

.payload h2 {
  margin: 0 0 0.5rem;
  font-size: 1rem;
}

pre {
  margin: 0;
  padding: 0.9rem 1rem;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 10px;
  overflow: auto;
  font-size: 0.82rem;
  line-height: 1.45;
  max-height: 28rem;
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

@media (max-width: 640px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
