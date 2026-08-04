<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { LoginUseCase } from '../application/auth-use-cases'
import { HttpAuthRepository } from '../infrastructure/http-auth-repository'
import { LocalStorageSessionStore } from '../infrastructure/local-storage-session-store'

const router = useRouter()
const email = ref('shop@ecomerce.local')
const password = ref('')
const error = ref('')
const loading = ref(false)

const login = new LoginUseCase(new HttpAuthRepository(), new LocalStorageSessionStore())

async function onSubmit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    await login.execute(email.value.trim(), password.value)
    await router.replace('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <form class="card" @submit.prevent="onSubmit" aria-label="Merchant login">
      <h1>Ecomerce Merchant</h1>
      <p class="lead">Đăng nhập để quản lý cửa hàng</p>

      <label>
        Email
        <input v-model="email" type="email" autocomplete="username" required />
      </label>

      <label>
        Password
        <input v-model="password" type="password" autocomplete="current-password" required />
      </label>

      <p v-if="error" class="error" role="alert">{{ error }}</p>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Đang đăng nhập…' : 'Đăng nhập' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background:
    linear-gradient(160deg, #0c4a6e 0%, #0369a1 45%, #0e7490 100%);
  font-family: 'IBM Plex Sans', 'Segoe UI', system-ui, sans-serif;
}

.card {
  width: min(420px, 100%);
  background: #fff;
  border-radius: 14px;
  padding: 1.6rem 1.5rem 1.4rem;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  box-shadow: 0 20px 50px rgba(8, 47, 73, 0.35);
}

h1 {
  margin: 0;
  font-size: 1.45rem;
  color: #0c4a6e;
}

.lead {
  margin: 0 0 0.35rem;
  color: #64748b;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.9rem;
  color: #334155;
  font-weight: 600;
}

input {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 0.65rem 0.75rem;
  font: inherit;
  font-weight: 400;
}

input:focus {
  outline: 2px solid #0369a1;
  outline-offset: 1px;
}

button {
  margin-top: 0.35rem;
  border: 0;
  border-radius: 8px;
  background: #0c4a6e;
  color: #fff;
  padding: 0.7rem 1rem;
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}

button:disabled {
  opacity: 0.7;
  cursor: wait;
}

.error {
  margin: 0;
  color: #b91c1c;
  font-size: 0.9rem;
}
</style>
