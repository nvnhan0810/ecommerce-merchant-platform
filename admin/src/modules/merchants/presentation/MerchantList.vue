<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { ListMerchantsUseCase } from '../application/list-merchants'
import { HttpAdminRepository } from '../../dashboard/infrastructure/http-admin-repository'

const useCase = new ListMerchantsUseCase(new HttpAdminRepository())

const { data, isLoading, isError, error } = useQuery({
  queryKey: ['admin', 'merchants'],
  queryFn: () => useCase.execute(),
})
</script>

<template>
  <section class="panel">
    <h1>Merchants</h1>
    <p v-if="isLoading">Đang tải…</p>
    <p v-else-if="isError" class="error">{{ (error as Error).message }}</p>
    <ul v-else-if="data" class="list" aria-label="Danh sách merchant">
      <li v-for="merchant in data" :key="merchant.id">
        <strong>{{ merchant.displayName }}</strong>
        <span>{{ merchant.email }}</span>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.panel h1 {
  margin: 0 0 1rem;
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
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.85rem 1rem;
}

.list span {
  color: #64748b;
}

.error {
  color: #b91c1c;
}
</style>
