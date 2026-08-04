<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

export type CategoryTagOption = {
  id: string
  name: string
  status?: string
  statusLabel?: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    options: CategoryTagOption[]
    /** Extra labels for selected ids not present in options (e.g. rejected). */
    selectedOptions?: CategoryTagOption[]
    creating?: boolean
    allowCreate?: boolean
    placeholder?: string
    hint?: string
  }>(),
  {
    selectedOptions: () => [],
    creating: false,
    allowCreate: true,
    placeholder: 'Tìm hoặc tạo danh mục…',
    hint: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
  create: [name: string]
}>()

const query = ref('')
const open = ref(false)
const highlight = ref(0)
const rootEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)
const pendingCreateName = ref('')

function stripDiacritics(s: string): string {
  return s
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/đ/g, 'd')
    .replace(/Đ/g, 'D')
    .toLowerCase()
}

const optionById = computed(() => {
  const map = new Map<string, CategoryTagOption>()
  for (const o of props.selectedOptions) map.set(o.id, o)
  for (const o of props.options) map.set(o.id, o)
  return map
})

const selectedTags = computed(() =>
  props.modelValue
    .map((id) => optionById.value.get(id) ?? { id, name: id })
    .filter(Boolean),
)

const available = computed(() =>
  props.options.filter((o) => !props.modelValue.includes(o.id)),
)

const filtered = computed(() => {
  const q = stripDiacritics(query.value.trim())
  if (!q) return available.value.slice(0, 12)
  return available.value
    .filter((o) => stripDiacritics(o.name).includes(q))
    .slice(0, 12)
})

const exactMatch = computed(() => {
  const q = stripDiacritics(query.value.trim())
  if (!q) return null
  return (
    available.value.find((o) => stripDiacritics(o.name) === q) ??
    null
  )
})

const canCreate = computed(() => {
  if (!props.allowCreate) return false
  const name = query.value.trim()
  if (!name) return false
  if (exactMatch.value) return false
  // already selected with same name
  const selectedSame = selectedTags.value.some(
    (t) => stripDiacritics(t.name) === stripDiacritics(name),
  )
  return !selectedSame
})

const menuItemsCount = computed(
  () => filtered.value.length + (canCreate.value ? 1 : 0),
)

watch(filtered, () => {
  highlight.value = 0
})

watch(
  () => props.options,
  () => {
    const pending = pendingCreateName.value.trim()
    if (!pending) return
    const created = props.options.find(
      (o) => stripDiacritics(o.name) === stripDiacritics(pending),
    )
    if (created) {
      addTag(created.id)
      pendingCreateName.value = ''
      query.value = ''
    }
  },
  { deep: true },
)

function addTag(id: string): void {
  if (props.modelValue.includes(id)) return
  emit('update:modelValue', [...props.modelValue, id])
  query.value = ''
  open.value = false
  void nextTick(() => inputEl.value?.focus())
}

function removeTag(id: string): void {
  emit(
    'update:modelValue',
    props.modelValue.filter((x) => x !== id),
  )
}

function requestCreate(name: string): void {
  const trimmed = name.trim()
  if (!trimmed || props.creating) return
  pendingCreateName.value = trimmed
  emit('create', trimmed)
}

function onFocus(): void {
  open.value = true
}

function onInput(): void {
  open.value = true
  highlight.value = 0
}

function onEnter(e: KeyboardEvent): void {
  e.preventDefault()
  const name = query.value.trim()
  if (!name) return

  if (exactMatch.value) {
    addTag(exactMatch.value.id)
    return
  }

  if (filtered.value.length > 0 && highlight.value < filtered.value.length) {
    addTag(filtered.value[highlight.value]!.id)
    return
  }

  if (canCreate.value) {
    requestCreate(name)
  }
}

function onKeydown(e: KeyboardEvent): void {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    open.value = true
    if (menuItemsCount.value === 0) return
    highlight.value = (highlight.value + 1) % menuItemsCount.value
    return
  }
  if (e.key === 'ArrowUp') {
    e.preventDefault()
    open.value = true
    if (menuItemsCount.value === 0) return
    highlight.value =
      (highlight.value - 1 + menuItemsCount.value) % menuItemsCount.value
    return
  }
  if (e.key === 'Enter') {
    onEnter(e)
    return
  }
  if (e.key === 'Escape') {
    open.value = false
    return
  }
  if (e.key === 'Backspace' && !query.value && props.modelValue.length) {
    removeTag(props.modelValue[props.modelValue.length - 1]!)
  }
}

function onCreateClick(): void {
  requestCreate(query.value)
}

function onDocPointerDown(e: Event): void {
  const target = e.target as Node | null
  if (rootEl.value && target && !rootEl.value.contains(target)) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', onDocPointerDown, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocPointerDown, true)
})

function selectCategory(id: string): void {
  addTag(id)
}

defineExpose({ selectCategory })
</script>

<template>
  <div ref="rootEl" class="tag-input">
    <p v-if="hint" class="hint">{{ hint }}</p>
    <div
      class="control"
      :class="{ open, creating }"
      @click="inputEl?.focus()"
    >
      <span
        v-for="tag in selectedTags"
        :key="tag.id"
        class="tag"
        :data-status="tag.status || 'approved'"
      >
        <span class="tag-label">
          {{ tag.name }}
          <small v-if="tag.status && tag.status !== 'approved' && tag.statusLabel">
            ({{ tag.statusLabel }})
          </small>
        </span>
        <button
          type="button"
          class="tag-remove"
          :aria-label="`Gỡ ${tag.name}`"
          @click.stop="removeTag(tag.id)"
        >
          ×
        </button>
      </span>
      <input
        ref="inputEl"
        v-model="query"
        type="text"
        class="search"
        :placeholder="selectedTags.length ? '' : placeholder"
        maxlength="120"
        autocomplete="off"
        @focus="onFocus"
        @input="onInput"
        @keydown="onKeydown"
      />
    </div>

    <ul v-if="open && (filtered.length || canCreate)" class="menu" role="listbox">
      <li
        v-for="(opt, idx) in filtered"
        :key="opt.id"
        class="option"
        :class="{ active: highlight === idx }"
        role="option"
        @mousedown.prevent="addTag(opt.id)"
        @mouseenter="highlight = idx"
      >
        <span>{{ opt.name }}</span>
        <span v-if="opt.statusLabel" class="badge" :data-status="opt.status">
          {{ opt.statusLabel }}
        </span>
      </li>
      <li
        v-if="canCreate"
        class="option create"
        :class="{ active: highlight === filtered.length }"
        role="option"
        @mousedown.prevent="onCreateClick"
        @mouseenter="highlight = filtered.length"
      >
        <span>
          {{ creating ? 'Đang tạo…' : 'Tạo mới' }}:
          <strong>{{ query.trim() }}</strong>
        </span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.tag-input {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.hint {
  margin: 0;
  color: #64748b;
  font-size: 0.85rem;
  font-weight: 400;
}

.control {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  min-height: 2.6rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.35rem 0.5rem;
  background: #fff;
  cursor: text;
}

.control.open {
  border-color: #94a3b8;
  box-shadow: 0 0 0 3px rgba(15, 23, 42, 0.06);
}

.control.creating {
  opacity: 0.85;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  max-width: 100%;
  padding: 0.15rem 0.25rem 0.15rem 0.5rem;
  border-radius: 999px;
  background: #f1f5f9;
  font-size: 0.85rem;
  font-weight: 500;
  color: #0f172a;
}

.tag[data-status='pending'] {
  background: #fef3c7;
  color: #92400e;
}

.tag[data-status='rejected'] {
  background: #fee2e2;
  color: #991b1b;
}

.tag-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag-label small {
  font-weight: 400;
  opacity: 0.85;
}

.tag-remove {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-size: 1.05rem;
  line-height: 1;
  padding: 0 0.25rem;
  border-radius: 4px;
  opacity: 0.7;
}

.tag-remove:hover {
  opacity: 1;
  background: rgba(15, 23, 42, 0.08);
}

.search {
  flex: 1;
  min-width: 8rem;
  border: 0 !important;
  outline: none;
  padding: 0.35rem 0.25rem !important;
  font: inherit;
  font-weight: 400;
  background: transparent;
}

.menu {
  position: absolute;
  z-index: 30;
  left: 0;
  right: 0;
  top: calc(100% + 4px);
  margin: 0;
  padding: 0.35rem;
  list-style: none;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.12);
  max-height: 240px;
  overflow: auto;
}

.option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.55rem 0.65rem;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.92rem;
  font-weight: 400;
  color: #0f172a;
}

.option.active,
.option:hover {
  background: #f1f5f9;
}

.option.create {
  border-top: 1px solid #e2e8f0;
  margin-top: 0.2rem;
  color: #0f172a;
}

.badge {
  font-size: 0.72rem;
  padding: 0.05rem 0.4rem;
  border-radius: 999px;
  background: #e2e8f0;
  color: #475569;
  flex-shrink: 0;
}

.badge[data-status='approved'] {
  background: #dcfce7;
  color: #166534;
}

.badge[data-status='pending'] {
  background: #fef3c7;
  color: #92400e;
}

.badge[data-status='rejected'] {
  background: #fee2e2;
  color: #991b1b;
}
</style>
