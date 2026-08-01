<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { ListMerchantProductsUseCase } from '../application/list-products'
import { HttpMerchantProductRepository } from '../infrastructure/http-product-repository'

const useCase = new ListMerchantProductsUseCase(new HttpMerchantProductRepository())

const { data, isLoading, isError, error } = useQuery({
  queryKey: ['merchant', 'products'],
  queryFn: () => useCase.execute(),
})
</script>

<template>
  <section class="panel">
    <h1>Sản phẩm của shop</h1>
    <p v-if="isLoading">Đang tải…</p>
    <p v-else-if="isError" class="error">{{ (error as Error).message }}</p>
    <ul v-else-if="data" class="list" aria-label="Sản phẩm merchant">
      <li v-for="product in data" :key="product.id.value">
        <div>
          <strong>{{ product.name }}</strong>
          <p>{{ product.description }}</p>
        </div>
        <div class="meta">
          <span>{{ product.price.format() }}</span>
          <span>Kho: {{ product.stock }}</span>
        </div>
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
  gap: 0.75rem;
}

.list li {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.9rem 1rem;
}

.list p {
  margin: 0.25rem 0 0;
  color: #64748b;
  font-size: 0.9rem;
}

.meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
  white-space: nowrap;
  color: #0f172a;
  font-weight: 600;
}

.error {
  color: #b91c1c;
}
</style>
