import type { JSX } from 'react'
import { ProductCatalog } from '@/modules/catalog/presentation/ProductCatalog'
import styles from './HomePage.module.css'

export function HomePage(): JSX.Element {
  return (
    <section className={styles.page}>
      <div className={styles.hero}>
        <h1>Ecomerce</h1>
        <p className={styles.lead}>Chọn sản phẩm, thêm vào giỏ và đặt hàng sau khi đăng nhập.</p>
      </div>
      <ProductCatalog />
    </section>
  )
}
