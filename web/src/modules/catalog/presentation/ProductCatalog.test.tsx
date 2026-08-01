import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { Money, Product, ProductId } from '../domain/product'
import { ProductCatalog } from './ProductCatalog'

vi.mock('../application/list-products', () => {
  return {
    ListProductsUseCase: class {
      execute = vi.fn(async () => [
        new Product(
          new ProductId('p1'),
          'Áo thun basic',
          'Cotton 100%',
          new Money(149000, 'VND'),
          50,
          'merchant-demo',
        ),
      ])
    },
  }
})

vi.mock('../infrastructure/http-product-repository', () => ({
  HttpProductRepository: class {},
}))

function renderCatalog(): void {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <ProductCatalog />
    </QueryClientProvider>,
  )
}

describe('ProductCatalog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should_render_product_name_when_loaded', async () => {
    renderCatalog()
    expect(await screen.findByText('Áo thun basic')).toBeInTheDocument()
    expect(screen.getByText(/Cotton 100%/)).toBeInTheDocument()
  })
})
