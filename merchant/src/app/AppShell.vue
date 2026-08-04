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
      <RouterLink class="brand" to="/">Ecomerce Merchant</RouterLink>
      <nav aria-label="Merchant">
        <RouterLink to="/">Dashboard</RouterLink>
        <RouterLink to="/products">Sản phẩm</RouterLink>
        <RouterLink to="/orders">Đơn hàng</RouterLink>
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
  height: 100vh;
  height: 100dvh;
  display: grid;
  grid-template-columns: 240px 1fr;
  overflow: hidden;
  background:
    linear-gradient(160deg, #f0f9ff 0%, #f8fafc 45%, #ecfeff 100%);
  color: #0f172a;
  font-family: 'IBM Plex Sans', 'Segoe UI', system-ui, sans-serif;
}

aside {
  background: #0c4a6e;
  color: #e0f2fe;
  padding: 1.25rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  min-height: 0;
  overflow-y: auto;
}

.brand {
  color: #fff;
  text-decoration: none;
  font-weight: 800;
  letter-spacing: -0.02em;
  flex-shrink: 0;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
  min-height: 0;
}

nav a {
  color: #bae6fd;
  text-decoration: none;
  padding: 0.45rem 0.6rem;
  border-radius: 8px;
}

nav a.router-link-exact-active,
nav a:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex-shrink: 0;
  margin-top: auto;
}

.user {
  margin: 0;
  font-size: 0.85rem;
  color: #7dd3fc;
}

button {
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: transparent;
  color: #e0f2fe;
  border-radius: 8px;
  padding: 0.4rem 0.6rem;
  cursor: pointer;
  font: inherit;
}

button:hover {
  background: rgba(255, 255, 255, 0.1);
}

main {
  padding: 1.75rem 1.5rem;
  min-height: 0;
  overflow-y: auto;
}

@media (max-width: 800px) {
  .shell {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }

  aside {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    overflow: visible;
  }

  nav {
    flex-direction: row;
    flex: initial;
  }

  .footer {
    margin-top: 0;
  }
}
</style>
