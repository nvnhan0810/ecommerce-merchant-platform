import { useState, type FormEvent, type JSX } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { CreateOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import styles from './CheckoutPage.module.css'

const createOrder = new CreateOrderUseCase(new HttpOrderRepository())

function formatMoney(cents: number, currency: string): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: currency || 'VND',
    maximumFractionDigits: 0,
  }).format(cents)
}

export function CheckoutPage(): JSX.Element {
  const { session } = useAuth()
  const { cart, clear } = useCart()
  const navigate = useNavigate()
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  if (!session?.isUser) {
    return <Navigate to="/login" replace state={{ from: '/checkout' }} />
  }

  if (cart.itemCount === 0) {
    return (
      <section className={styles.page}>
        <h1>Đặt hàng</h1>
        <p>
          Giỏ hàng trống. <Link to="/">Chọn sản phẩm</Link>
        </p>
      </section>
    )
  }

  async function onSubmit(e: FormEvent): Promise<void> {
    e.preventDefault()
    setError('')
    setPending(true)
    try {
      const orders = await createOrder.execute(
        note,
        cart.items.map((item) => ({ productId: item.productId, quantity: item.quantity })),
      )
      clear()
      if (orders.length === 1) {
        void navigate(`/orders/${orders[0].id}`)
      } else {
        void navigate('/orders')
      }
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <section className={styles.page}>
      <h1>Đặt hàng</h1>
      <p className={styles.meta}>Đăng nhập với {session.email}</p>
      <ul className={styles.list}>
        {cart.items.map((item) => (
          <li key={item.productId}>
            <span>
              {item.name} × {item.quantity}
            </span>
            <strong>{formatMoney(item.lineTotalCents, item.currency)}</strong>
          </li>
        ))}
      </ul>
      <p className={styles.total}>
        Tổng: <strong>{formatMoney(cart.totalCents, cart.items[0]?.currency || 'VND')}</strong>
      </p>
      <form className={styles.form} onSubmit={(e) => void onSubmit(e)}>
        <label>
          Ghi chú
          <textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} />
        </label>
        {error && <p className={styles.error}>{error}</p>}
        <button type="submit" disabled={pending}>
          {pending ? 'Đang đặt…' : 'Xác nhận đặt hàng'}
        </button>
      </form>
    </section>
  )
}
