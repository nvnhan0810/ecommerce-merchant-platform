<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { GetAdminOverviewUseCase } from '../application/get-overview'
import { HttpAdminRepository } from '../infrastructure/http-admin-repository'

const useCase = new GetAdminOverviewUseCase(new HttpAdminRepository())

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['admin', 'overview'],
  queryFn: () => useCase.execute(),
})
</script>

<template>
  <section class="panel">
    <header>
      <h1>Admin console</h1>
      <p>ecomerce-admin.nvnhan0810.com</p>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" @click="refetch()">Thử lại</button>
    </div>
    <div v-else-if="data" class="stats" aria-label="Tổng quan hệ thống">
      <article>
        <span>Users</span>
        <strong>{{ data.userCount }}</strong>
      </article>
      <article>
        <span>Merchants</span>
        <strong>{{ data.merchantCount }}</strong>
      </article>
      <article>
        <span>Tổng tài khoản</span>
        <strong>{{ data.totalAccounts }}</strong>
      </article>
    </div>
  </section>
</template>

<style scoped>
.panel {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

header h1 {
  margin: 0;
}

header p {
  margin: 0.35rem 0 0;
  color: #64748b;
}

.stats {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.stats article {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.stats span {
  color: #64748b;
  font-size: 0.85rem;
}

.stats strong {
  font-size: 1.35rem;
}

.error {
  color: #b91c1c;
}

button {
  margin-top: 0.5rem;
  border: 0;
  background: #334155;
  color: #fff;
  border-radius: 8px;
  padding: 0.4rem 0.85rem;
  cursor: pointer;
}
</style>
