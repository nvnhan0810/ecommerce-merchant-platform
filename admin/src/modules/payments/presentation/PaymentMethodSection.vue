<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    id: string
    title: string
    description?: string
    badge?: string
    badgeTone?: 'ready' | 'pending' | 'muted'
    defaultOpen?: boolean
  }>(),
  {
    description: '',
    badge: '',
    badgeTone: 'muted',
    defaultOpen: false,
  },
)

const open = ref(props.defaultOpen)

watch(
  () => props.defaultOpen,
  (value) => {
    open.value = value
  },
)

const panelId = computed(() => `payment-section-${props.id}`)

function toggle(): void {
  open.value = !open.value
}
</script>

<template>
  <section class="section" :class="{ open }">
    <button
      type="button"
      class="header"
      :aria-expanded="open"
      :aria-controls="panelId"
      @click="toggle"
    >
      <span class="chevron" aria-hidden="true">{{ open ? '▾' : '▸' }}</span>
      <span class="titles">
        <strong>{{ title }}</strong>
        <small v-if="description">{{ description }}</small>
      </span>
      <span v-if="badge" class="badge" :class="badgeTone">{{ badge }}</span>
    </button>
    <div v-show="open" :id="panelId" class="body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.section {
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #fff;
  overflow: hidden;
}

.header {
  width: 100%;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border: 0;
  background: #f8fafc;
  color: inherit;
  text-align: left;
  padding: 0.95rem 1rem;
  cursor: pointer;
  font: inherit;
}

.section.open .header {
  border-bottom: 1px solid #e2e8f0;
}

.chevron {
  width: 1rem;
  flex-shrink: 0;
  margin-top: 0.15rem;
  color: #64748b;
  font-size: 0.9rem;
}

.titles {
  display: grid;
  gap: 0.2rem;
  flex: 1;
  min-width: 0;
}

.titles strong {
  font-size: 1rem;
}

.titles small {
  color: #64748b;
  font-size: 0.86rem;
  line-height: 1.4;
}

.badge {
  flex-shrink: 0;
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.3;
}

.badge.ready {
  background: #ecfdf5;
  color: #047857;
}

.badge.pending {
  background: #fff7ed;
  color: #c2410c;
}

.badge.muted {
  background: #e2e8f0;
  color: #475569;
}

.body {
  padding: 1rem;
  display: grid;
  gap: 0.9rem;
}
</style>
