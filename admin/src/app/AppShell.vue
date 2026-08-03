<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { GetStoredSessionUseCase, LogoutUseCase } from '@/modules/auth/application/auth-use-cases'
import { LocalStorageSessionStore } from '@/modules/auth/infrastructure/local-storage-session-store'

const router = useRouter()
const store = new LocalStorageSessionStore()
const session = computed(() => new GetStoredSessionUseCase(store).execute())

function logout(): void {
  new LogoutUseCase(store).execute()
  void router.replace('/login')
}
</script>

<template>
  <div class="shell">
    <aside>
      <RouterLink class="brand" to="/">Ecomerce Admin</RouterLink>
      <nav aria-label="Admin">
        <RouterLink to="/">Overview</RouterLink>
        <RouterLink to="/users">Users</RouterLink>
        <RouterLink to="/merchants">Merchants</RouterLink>
        <RouterLink to="/products">Products</RouterLink>
      </nav>
      <div class="footer">
        <p v-if="session" class="user">{{ session.displayName }}</p>
        <button type="button" @click="logout">Đăng xuất</button>
      </div>
    </aside>
    <main>
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 240px 1fr;
  background:
    linear-gradient(145deg, #f8fafc 0%, #e2e8f0 50%, #f1f5f9 100%);
  color: #0f172a;
  font-family: 'IBM Plex Sans', 'Segoe UI', system-ui, sans-serif;
}

aside {
  background: #1e293b;
  color: #e2e8f0;
  padding: 1.25rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.brand {
  color: #fff;
  text-decoration: none;
  font-weight: 800;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

nav a {
  color: #cbd5e1;
  text-decoration: none;
  padding: 0.45rem 0.6rem;
  border-radius: 8px;
}

nav a.router-link-exact-active,
nav a:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.user {
  margin: 0;
  font-size: 0.85rem;
  color: #94a3b8;
}

button {
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  color: #e2e8f0;
  border-radius: 8px;
  padding: 0.4rem 0.6rem;
  cursor: pointer;
  font: inherit;
}

button:hover {
  background: rgba(255, 255, 255, 0.08);
}

main {
  padding: 1.75rem 1.5rem;
}

@media (max-width: 800px) {
  .shell {
    grid-template-columns: 1fr;
  }

  aside {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
  }

  nav {
    flex-direction: row;
    flex: initial;
  }
}
</style>
