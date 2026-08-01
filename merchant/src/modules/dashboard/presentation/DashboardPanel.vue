<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { GetMerchantDashboardUseCase } from '../application/get-dashboard'
import { HttpDashboardRepository } from '../infrastructure/http-dashboard-repository'

const useCase = new GetMerchantDashboardUseCase(new HttpDashboardRepository())

const { data, isLoading, isError, error, refetch } = useQuery({
  queryKey: ['merchant', 'dashboard'],
  queryFn: () => useCase.execute(),
})
</script>

<template>
  <section class="panel">
    <header>
      <h1>Merchant dashboard</h1>
      <p>ecomerce-merchant.nvnhan0810.com</p>
    </header>

    <p v-if="isLoading">Đang tải…</p>
    <div v-else-if="isError" class="error">
      <p>{{ (error as Error).message }}</p>
      <button type="button" @click="refetch()">Thử lại</button>
    </div>
    <div v-else-if="data" class="stats" aria-label="Thống kê">
      <article>
        <span>Sản phẩm</span>
        <strong>{{ data.productCount }}</strong>
      </article>
      <article>
        <span>Đơn hàng</span>
        <strong>{{ data.orderCount }}</strong>
      </article>
      <article>
        <span>Giá trị catalog</span>
        <strong>{{ data.revenueLabel }}</strong>
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
  font-size: 1.75rem;
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
  color: #0f172a;
}

.error {
  color: #b91c1c;
}

button {
  margin-top: 0.5rem;
  border: 0;
  background: #0369a1;
  color: #fff;
  border-radius: 8px;
  padding: 0.4rem 0.85rem;
  cursor: pointer;
}
</style>
