<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="shell">
    <aside>
      <RouterLink class="brand" to="/">Ecomerce Merchant</RouterLink>
      <nav aria-label="Merchant">
        <RouterLink to="/">Dashboard</RouterLink>
        <RouterLink to="/products">Sản phẩm</RouterLink>
      </nav>
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
}

.brand {
  color: #fff;
  text-decoration: none;
  font-weight: 800;
  letter-spacing: -0.02em;
}

nav {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

nav a {
  color: #bae6fd;
  text-decoration: none;
  padding: 0.45rem 0.6rem;
  border-radius: 8px;
}

nav a.router-link-active,
nav a:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
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
  }

  nav {
    flex-direction: row;
  }
}
</style>
