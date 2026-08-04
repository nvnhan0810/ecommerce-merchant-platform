import type { JSX } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { ListMyOrdersUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import { OrderStatusBadge } from './OrderStatusBadge'
import styles from './OrdersPage.module.css'

const listOrders = new ListMyOrdersUseCase(new HttpOrderRepository())

function formatMoney(cents: number, currency: string): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: currency || 'VND',
    maximumFractionDigits: 0,
  }).format(cents)
}

export function OrdersPage(): JSX.Element {
  const { session } = useAuth()
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['user', 'orders'],
    queryFn: () => listOrders.execute(),
    enabled: Boolean(session?.isUser),
  })

  if (!session?.isUser) {
    return <Navigate to="/login" replace state={{ from: '/orders' }} />
  }

  return (
    <section className={styles.page}>
      <h1>Lịch sử đặt hàng</h1>
      {isLoading && <p>Đang tải…</p>}
      {isError && (
        <div>
          <p>{(error as Error).message}</p>
          <button type="button" onClick={() => void refetch()}>
            Thử lại
          </button>
        </div>
      )}
      {data && (
        <ul className={styles.list} aria-label="Đơn hàng của tôi">
          {data.map((order) => (
            <li key={order.id}>
              <Link to={`/orders/${order.id}`} className={styles.item}>
                <div>
                  <strong className={styles.code}>{order.code}</strong>
                  <div className={styles.meta}>
                    <OrderStatusBadge status={order.status} label={order.statusLabel} />
                    <span>· {formatMoney(order.totalCents, order.currency)}</span>
                  </div>
                </div>
                <span className={styles.viewBtn}>Xem chi tiết</span>
              </Link>
            </li>
          ))}
          {data.length === 0 && <li className={styles.empty}>Chưa có đơn hàng.</li>}
        </ul>
      )}
    </section>
  )
}
