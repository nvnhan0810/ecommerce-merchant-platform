import type { JSX } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ListProductsUseCase } from '../application/list-products'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import styles from './ProductCatalog.module.css'

const listProducts = new ListProductsUseCase(new HttpProductRepository())

export function ProductCatalog(): JSX.Element {
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['catalog', 'products'],
    queryFn: () => listProducts.execute(20),
  })

  if (isLoading) {
    return <p className={styles.status}>Đang tải sản phẩm…</p>
  }

  if (isError) {
    return (
      <div className={styles.status}>
        <p>Không tải được danh mục: {(error as Error).message}</p>
        <button type="button" onClick={() => void refetch()}>
          Thử lại
        </button>
      </div>
    )
  }

  if (!data?.length) {
    return <p className={styles.status}>Chưa có sản phẩm.</p>
  }

  return (
    <ul className={styles.grid} aria-label="Danh sách sản phẩm">
      {data.map((product) => (
        <li key={product.id.value} className={styles.card}>
          <h3>{product.name}</h3>
          <p className={styles.desc}>{product.description}</p>
          <div className={styles.meta}>
            <span className={styles.price}>{product.price.format()}</span>
            <span className={product.isAvailable ? styles.inStock : styles.out}>
              {product.isAvailable ? `Còn ${product.stock}` : 'Hết hàng'}
            </span>
          </div>
        </li>
      ))}
    </ul>
  )
}
