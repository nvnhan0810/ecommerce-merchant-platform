import type { JSX } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { GetMyOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import styles from './OrderDetailPage.module.css'

const getOrder = new GetMyOrderUseCase(new HttpOrderRepository())

function formatMoney(cents: number, currency: string): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: currency || 'VND',
    maximumFractionDigits: 0,
  }).format(cents)
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('vi-VN')
  } catch {
    return iso
  }
}

export function OrderDetailPage(): JSX.Element {
  const { session } = useAuth()
  const { id = '' } = useParams()
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['user', 'orders', id],
    queryFn: () => getOrder.execute(id),
    enabled: Boolean(session?.isUser && id),
  })

  if (!session?.isUser) {
    return <Navigate to="/login" replace state={{ from: `/orders/${id}` }} />
  }

  return (
    <section className={styles.page}>
      <Link className={styles.back} to="/orders">
        ← Lịch sử đặt hàng
      </Link>
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
        <article className={styles.card}>
          <h1 className={styles.code}>{data.code}</h1>
          <p>
            Trạng thái: <strong>{data.statusLabel}</strong>
          </p>
          <p>
            Mã vận đơn: <strong className={styles.codeInline}>{data.deliveryTrackingCode || '—'}</strong>
          </p>
          <p>Đơn vị vận chuyển: {data.deliveryCarrier || 'internal'}</p>
          <p>Tổng: {formatMoney(data.totalCents, data.currency)}</p>
          <p>Tạo lúc: {formatDate(data.createdAt)}</p>
          <p>Ghi chú: {data.note || '—'}</p>
          <h2>Sản phẩm</h2>
          <table>
            <thead>
              <tr>
                <th>Tên</th>
                <th>SL</th>
                <th>Thành tiền</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <tr key={item.id}>
                  <td>{item.productName}</td>
                  <td>{item.quantity}</td>
                  <td>{formatMoney(item.lineTotalCents, data.currency)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <h2>Timeline vận chuyển</h2>
          {data.deliveryEvents.length === 0 ? (
            <p className={styles.muted}>Chưa có sự kiện vận chuyển.</p>
          ) : (
            <ol className={styles.timeline}>
              {data.deliveryEvents.map((ev) => (
                <li key={ev.id}>
                  <strong>{ev.statusLabel || ev.statusCode}</strong>
                  <span className={styles.muted}>{formatDate(ev.occurredAt)}</span>
                  <p>{ev.message}</p>
                  {ev.reason ? <p className={styles.muted}>Lý do: {ev.reason}</p> : null}
                </li>
              ))}
            </ol>
          )}
        </article>
      )}
    </section>
  )
}
