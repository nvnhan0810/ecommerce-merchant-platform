import type { JSX } from 'react'
import { Cart } from '@/modules/cart/domain/cart'
import styles from './CartPage.module.css'

export function CartPage(): JSX.Element {
  const cart = new Cart([])

  return (
    <section className={styles.page}>
      <h1>Giỏ hàng</h1>
      <p>
        {cart.itemCount === 0
          ? 'Giỏ hàng trống — thêm sản phẩm từ trang chủ (sắp có).'
          : `${cart.itemCount} sản phẩm`}
      </p>
    </section>
  )
}
