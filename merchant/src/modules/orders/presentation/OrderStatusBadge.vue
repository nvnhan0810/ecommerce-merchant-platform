<script setup lang="ts">
import type { OrderStatus } from '../domain/order'
import { ORDER_STATUS_OPTIONS } from '../domain/order'

const props = defineProps<{
  status: OrderStatus | string
  label?: string
}>()

const resolvedLabel = (): string => {
  if (props.label) return props.label
  return ORDER_STATUS_OPTIONS.find((o) => o.value === props.status)?.label ?? String(props.status)
}
</script>

<template>
  <span class="badge" :class="`badge--${status}`">{{ resolvedLabel() }}</span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 650;
  letter-spacing: 0.01em;
  border: 1px solid transparent;
  white-space: nowrap;
}

.badge--new {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

.badge--awaiting_payment {
  background: #fff7ed;
  color: #c2410c;
  border-color: #fed7aa;
}

.badge--paid {
  background: #ecfeff;
  color: #0e7490;
  border-color: #a5f3fc;
}

.badge--confirmed {
  background: #f5f3ff;
  color: #6d28d9;
  border-color: #ddd6fe;
}

.badge--shipping {
  background: #fff7ed;
  color: #c2410c;
  border-color: #fed7aa;
}

.badge--succeeded {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

.badge--returning {
  background: #fffbeb;
  color: #b45309;
  border-color: #fde68a;
}

.badge--returned {
  background: #f8fafc;
  color: #475569;
  border-color: #cbd5e1;
}

.badge--failed {
  background: #fef2f2;
  color: #b91c1c;
  border-color: #fecaca;
}

.badge--cancelled {
  background: #f8fafc;
  color: #475569;
  border-color: #cbd5e1;
}
</style>
