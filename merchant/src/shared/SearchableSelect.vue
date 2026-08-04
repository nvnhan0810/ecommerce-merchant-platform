<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

export type SearchableOption = {
  value: string
  label: string
}

const props = withDefaults(
  defineProps<{
    options: SearchableOption[]
    modelValue: string
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
    required?: boolean
    ariaLabel?: string
  }>(),
  {
    placeholder: 'Tìm và chọn…',
    disabled: false,
    clearable: true,
    required: false,
    ariaLabel: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const query = ref('')
const rootEl = ref<HTMLElement | null>(null)
const controlEl = ref<HTMLElement | null>(null)
const menuEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)
const menuStyle = ref<Record<string, string>>({})

type TapState = { x: number; y: number; option: SearchableOption }
const tapState = ref<TapState | null>(null)

function stripDiacritics(s: string): string {
  return s
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/đ/g, 'd')
    .replace(/Đ/g, 'D')
    .toLowerCase()
}

const selected = computed(() => props.options.find((o) => o.value === props.modelValue) ?? null)

const filtered = computed(() => {
  const q = stripDiacritics(query.value.trim())
  if (!q) return props.options
  return props.options.filter((o) => {
    const hay = `${stripDiacritics(o.label)} ${stripDiacritics(o.value)}`
    return hay.includes(q)
  })
})

function isCoarsePointer(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches
}

function updateMenuPosition(): void {
  const control = controlEl.value
  if (!control) return
  const rect = control.getBoundingClientRect()
  const viewportH = window.innerHeight
  const viewportW = window.innerWidth
  const gap = 4
  const maxMenuH = Math.min(320, viewportH * 0.45)
  const spaceBelow = viewportH - rect.bottom - gap
  const spaceAbove = rect.top - gap
  const openUp = spaceBelow < 180 && spaceAbove > spaceBelow
  const height = Math.max(160, Math.min(maxMenuH, openUp ? spaceAbove : spaceBelow))

  menuStyle.value = {
    position: 'fixed',
    left: `${Math.max(8, Math.min(rect.left, viewportW - rect.width - 8))}px`,
    width: `${rect.width}px`,
    zIndex: '1000',
    maxHeight: `${height}px`,
    ...(openUp
      ? { bottom: `${viewportH - rect.top + gap}px`, top: 'auto' }
      : { top: `${rect.bottom + gap}px`, bottom: 'auto' }),
  }
}

async function openMenu(): Promise<void> {
  open.value = true
  query.value = ''
  tapState.value = null
  await nextTick()
  updateMenuPosition()
  if (!isCoarsePointer()) {
    inputEl.value?.focus()
  }
}

function closeMenu(): void {
  open.value = false
  query.value = ''
  tapState.value = null
}

async function toggle(): Promise<void> {
  if (props.disabled) return
  if (open.value) {
    closeMenu()
    return
  }
  await openMenu()
}

function selectOption(option: SearchableOption): void {
  emit('update:modelValue', option.value)
  closeMenu()
}

function clearValue(event: Event): void {
  event.preventDefault()
  event.stopPropagation()
  if (props.disabled || !props.clearable) return
  emit('update:modelValue', '')
  closeMenu()
}

function onOptionPointerDown(option: SearchableOption, event: PointerEvent): void {
  if (event.pointerType === 'mouse' && event.button !== 0) return
  tapState.value = { x: event.clientX, y: event.clientY, option }
}

function onOptionPointerMove(event: PointerEvent): void {
  const tap = tapState.value
  if (!tap) return
  const dx = Math.abs(event.clientX - tap.x)
  const dy = Math.abs(event.clientY - tap.y)
  // Treat as scroll/drag — do not select on release.
  if (dx > 8 || dy > 8) {
    tapState.value = null
  }
}

function onOptionPointerUp(event: PointerEvent): void {
  const tap = tapState.value
  tapState.value = null
  if (!tap) return
  const dx = Math.abs(event.clientX - tap.x)
  const dy = Math.abs(event.clientY - tap.y)
  if (dx <= 8 && dy <= 8) {
    event.preventDefault()
    selectOption(tap.option)
  }
}

function onOptionPointerCancel(): void {
  tapState.value = null
}

function onDocPointerDown(event: PointerEvent): void {
  if (!open.value) return
  const target = event.target as Node | null
  if (!target) return
  if (rootEl.value?.contains(target) || menuEl.value?.contains(target)) return
  // Close only — never emit a value change.
  closeMenu()
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') closeMenu()
}

function onViewportChange(): void {
  if (open.value) updateMenuPosition()
}

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) closeMenu()
  },
)

watch(
  () => props.options,
  () => {
    if (open.value) void nextTick(() => updateMenuPosition())
  },
)

onMounted(() => {
  // Bubble phase so option handlers run first; outside closes without changing value.
  document.addEventListener('pointerdown', onDocPointerDown)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocPointerDown)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<template>
  <div ref="rootEl" class="searchSelect" :class="{ disabled, open }">
    <div
      ref="controlEl"
      class="control"
      role="combobox"
      :tabindex="disabled ? -1 : 0"
      :aria-label="ariaLabel"
      :aria-expanded="open"
      :aria-disabled="disabled"
      @click="toggle"
      @keydown.enter.prevent="toggle"
      @keydown.space.prevent="toggle"
    >
      <span v-if="selected" class="value">{{ selected.label }}</span>
      <span v-else class="placeholder">{{ placeholder }}</span>
      <span class="indicators">
        <button
          v-if="clearable && modelValue && !disabled"
          type="button"
          class="clear"
          aria-label="Xóa lựa chọn"
          @pointerdown.stop.prevent="clearValue"
          @click.stop.prevent="clearValue"
        >
          ×
        </button>
        <span class="chevron" aria-hidden="true">▾</span>
      </span>
    </div>

    <Teleport to="body">
      <div
        v-if="open"
        ref="menuEl"
        class="searchSelectMenu"
        role="listbox"
        :style="menuStyle"
        @pointerdown.stop
      >
        <input
          ref="inputEl"
          v-model="query"
          class="search"
          type="search"
          enterkeyhint="search"
          :placeholder="placeholder"
          autocomplete="off"
          @click.stop
          @pointerdown.stop
        />
        <ul class="options">
          <li v-if="filtered.length === 0" class="empty">Không có kết quả</li>
          <li
            v-for="option in filtered"
            :key="option.value"
            role="option"
            :aria-selected="option.value === modelValue"
            :class="{ active: option.value === modelValue }"
            @pointerdown.stop="onOptionPointerDown(option, $event)"
            @pointermove.stop="onOptionPointerMove"
            @pointerup.stop="onOptionPointerUp"
            @pointercancel.stop="onOptionPointerCancel"
          >
            {{ option.label }}
          </li>
        </ul>
      </div>
    </Teleport>

    <input
      v-if="required"
      class="nativeRequired"
      tabindex="-1"
      aria-hidden="true"
      required
      :value="modelValue"
      @change="() => undefined"
    />
  </div>
</template>

<style scoped>
.searchSelect {
  position: relative;
  font-weight: 400;
}

.control {
  width: 100%;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.45rem 0.55rem 0.45rem 0.7rem;
  background: #fff;
  font: inherit;
  text-align: left;
  cursor: pointer;
  color: #0f172a;
  box-sizing: border-box;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.control:hover {
  border-color: #94a3b8;
}

.open .control {
  border-color: #0f766e;
  box-shadow: 0 0 0 1px #0f766e;
}

.disabled .control {
  background: #f1f5f9;
  color: #64748b;
  cursor: not-allowed;
}

.value {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.placeholder {
  flex: 1;
  color: #94a3b8;
}

.indicators {
  display: inline-flex;
  align-items: center;
  gap: 0.15rem;
  color: #64748b;
  flex-shrink: 0;
}

.clear {
  border: 0;
  background: transparent;
  color: #64748b;
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  padding: 0.25rem 0.35rem;
  min-width: 32px;
  min-height: 32px;
}

.chevron {
  font-size: 0.75rem;
}

.nativeRequired {
  opacity: 0;
  height: 0;
  width: 0;
  position: absolute;
  pointer-events: none;
}
</style>

<style>
.searchSelectMenu {
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.12);
  overflow: hidden;
  -webkit-overflow-scrolling: touch;
}

.searchSelectMenu .search {
  width: 100%;
  flex: 0 0 auto;
  border: 0;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.75rem 0.85rem;
  font: inherit;
  font-size: 16px;
  outline: none;
  box-sizing: border-box;
}

.searchSelectMenu .options {
  list-style: none;
  margin: 0;
  padding: 0.25rem 0;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1 1 auto;
  min-height: 0;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior: contain;
  touch-action: pan-y;
}

.searchSelectMenu .options li {
  padding: 0.75rem 0.85rem;
  cursor: pointer;
  color: #0f172a;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
}

.searchSelectMenu .options li:active {
  background: #f0fdfa;
}

.searchSelectMenu .options li.active {
  background: #0f766e;
  color: #fff;
}

.searchSelectMenu .options li.empty {
  color: #64748b;
  cursor: default;
}

.searchSelectMenu .options li.empty:active {
  background: transparent;
}
</style>
