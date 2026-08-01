import type { JSX } from 'react'
import { ProductCatalog } from '@/modules/catalog/presentation/ProductCatalog'
import styles from './HomePage.module.css'

export function HomePage(): JSX.Element {
  return (
    <section className={styles.page}>
      <div className={styles.hero}>
        <p className={styles.kicker}>Storefront</p>
        <h1>Ecomerce</h1>
        <p className={styles.lead}>
          Mua sắm nhanh — danh mục đồng bộ từ API Go tại ecomerce-api.nvnhan0810.com
        </p>
      </div>
      <ProductCatalog />
    </section>
  )
}
