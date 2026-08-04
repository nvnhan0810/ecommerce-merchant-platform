import type { JSX } from 'react'
import { Link } from 'react-router-dom'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import styles from './CartPage.module.css'

function formatMoney(cents: number, currency: string): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: currency || 'VND',
    maximumFractionDigits: 0,
  }).format(cents)
}

export function CartPage(): JSX.Element {
  const { cart, setQuantity, removeItem } = useCart()

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <h1>Giỏ hàng</h1>
        {cart.itemCount > 0 && (
          <Link className={styles.checkout} to="/checkout">
            Đặt hàng
          </Link>
        )}
      </header>

      {cart.itemCount === 0 ? (
        <p className={styles.empty}>
          Giỏ hàng trống. <Link to="/">Xem sản phẩm</Link>
        </p>
      ) : (
        <>
          <ul className={styles.list} aria-label="Giỏ hàng">
            {cart.items.map((item) => (
              <li key={item.productId} className={styles.row}>
                <div>
                  <Link to={`/products/${item.productId}`} className={styles.name}>
                    {item.name}
                  </Link>
                  <p>{formatMoney(item.unitPriceCents, item.currency)}</p>
                </div>
                <div className={styles.actions}>
                  <label>
                    SL
                    <input
                      type="number"
                      min={1}
                      value={item.quantity}
                      onChange={(e) => setQuantity(item.productId, Number(e.target.value))}
                    />
                  </label>
                  <strong>{formatMoney(item.lineTotalCents, item.currency)}</strong>
                  <button type="button" className={styles.remove} onClick={() => removeItem(item.productId)}>
                    Xóa
                  </button>
                </div>
              </li>
            ))}
          </ul>
          <p className={styles.total}>
            Tổng: <strong>{formatMoney(cart.totalCents, cart.items[0]?.currency || 'VND')}</strong>
          </p>
        </>
      )}
    </section>
  )
}
