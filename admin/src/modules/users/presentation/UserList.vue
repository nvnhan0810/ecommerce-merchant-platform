<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  CreateUserUseCase,
  DeleteUserUseCase,
  ListUsersUseCase,
  UpdateUserUseCase,
} from '../application/user-use-cases'
import { HttpUserRepository } from '../infrastructure/http-user-repository'
import type { UserAccount } from '../domain/user'

const repo = new HttpUserRepository()
const listUseCase = new ListUsersUseCase(repo)
const createUseCase = new CreateUserUseCase(repo)
const updateUseCase = new UpdateUserUseCase(repo)
const deleteUseCase = new DeleteUserUseCase(repo)

const queryClient = useQueryClient()
const formError = ref('')
const editingId = ref<string | null>(null)
const showForm = ref(false)

const form = reactive({
  email: '',
  displayName: '',
  password: '',
})

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'users'],
  queryFn: () => listUseCase.execute(),
})

const isEditing = computed(() => editingId.value !== null)

function resetForm(): void {
  form.email = ''
  form.displayName = ''
  form.password = ''
  editingId.value = null
  formError.value = ''
  showForm.value = false
}

function openCreate(): void {
  resetForm()
  showForm.value = true
}

function openEdit(user: UserAccount): void {
  editingId.value = user.id
  form.email = user.email
  form.displayName = user.displayName
  form.password = ''
  formError.value = ''
  showForm.value = true
}

const saveMutation = useMutation({
  mutationFn: async () => {
    if (isEditing.value && editingId.value) {
      return updateUseCase.execute({
        id: editingId.value,
        email: form.email.trim(),
        displayName: form.displayName.trim(),
        password: form.password.trim() || undefined,
      })
    }
    return createUseCase.execute({
      email: form.email.trim(),
      displayName: form.displayName.trim(),
      password: form.password,
    })
  },
  onSuccess: async () => {
    resetForm()
    await queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
  },
  onError: (e: Error) => {
    formError.value = e.message
  },
})

const deleteMutation = useMutation({
  mutationFn: (id: string) => deleteUseCase.execute(id),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
  },
})

async function onSubmit(): Promise<void> {
  formError.value = ''
  await saveMutation.mutateAsync()
}

async function onDelete(user: UserAccount): Promise<void> {
  if (!window.confirm(`Xóa user "${user.displayName}"?`)) {
    return
  }
  await deleteMutation.mutateAsync(user.id)
}
</script>

<template>
  <section class="panel">
    <header class="header">
      <h1>Users</h1>
      <button type="button" class="primary" @click="openCreate">Thêm user</button>
    </header>

    <form v-if="showForm" class="form" @submit.prevent="onSubmit" aria-label="User form">
      <h2>{{ isEditing ? 'Sửa user' : 'Tạo user' }}</h2>
      <label>
        Tên hiển thị
        <input v-model="form.displayName" required />
      </label>
      <label>
        Email
        <input v-model="form.email" type="email" required autocomplete="off" />
      </label>
      <label>
        Password
        <input
          v-model="form.password"
          type="password"
          :required="!isEditing"
          :placeholder="isEditing ? 'Để trống nếu giữ password cũ' : ''"
          autocomplete="new-password"
        />
      </label>
      <p v-if="formError" class="error" role="alert">{{ formError }}</p>
      <div class="actions">
        <button type="submit" class="primary" :disabled="saveMutation.isPending.value">
          {{ saveMutation.isPending.value ? 'Đang lưu…' : 'Lưu' }}
        </button>
        <button type="button" class="ghost" @click="resetForm">Hủy</button>
      </div>
    </form>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" class="ghost" @click="refetch()">Thử lại</button>
    </div>
    <ul v-else-if="data" class="list" aria-label="Danh sách user">
      <li v-for="user in data" :key="user.id">
        <div>
          <strong>{{ user.displayName }}</strong>
          <p>{{ user.email }}</p>
        </div>
        <div class="row-actions">
          <button type="button" class="ghost" @click="openEdit(user)">Sửa</button>
          <button
            type="button"
            class="danger"
            :disabled="deleteMutation.isPending.value"
            @click="onDelete(user)"
          >
            Xóa
          </button>
        </div>
      </li>
      <li v-if="data.length === 0" class="empty">Chưa có user nào.</li>
    </ul>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.header h1 {
  margin: 0;
}

.form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form h2 {
  margin: 0;
  font-size: 1.05rem;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #334155;
}

input {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.55rem 0.7rem;
  font: inherit;
  font-weight: 400;
}

.actions,
.row-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

button {
  border-radius: 8px;
  padding: 0.45rem 0.85rem;
  font: inherit;
  cursor: pointer;
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

.danger {
  border: 1px solid #fecaca;
  background: #fff1f2;
  color: #b91c1c;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.list li {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.85rem 1rem;
}

.list p {
  margin: 0.2rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.empty {
  color: #64748b;
  justify-content: center;
}

.error {
  color: #b91c1c;
  margin: 0;
}
</style>
