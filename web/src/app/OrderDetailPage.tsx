import { useState, type JSX } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { GetMyOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import { OrderStatusBadge } from './OrderStatusBadge'
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
  const [activeTab, setActiveTab] = useState<'details' | 'tracking'>('details')
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
          <div className={styles.header}>
            <h1 className={styles.code}>{data.code}</h1>
            <OrderStatusBadge status={data.status} label={data.statusLabel} />
          </div>

          <div className={styles.infoGrid}>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Mã vận đơn</span>
              <strong className={styles.codeInline}>{data.deliveryTrackingCode || '—'}</strong>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Đơn vị vận chuyển</span>
              <span>{data.deliveryCarrier || 'internal'}</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Tổng tiền</span>
              <strong className={styles.totalPrice}>{formatMoney(data.totalCents, data.currency)}</strong>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Ngày tạo</span>
              <span>{formatDate(data.createdAt)}</span>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Ghi chú</span>
              <span>{data.note || '—'}</span>
            </div>
          </div>

          <div className={styles.tabs}>
            <button
              type="button"
              className={`${styles.tabBtn} ${activeTab === 'details' ? styles.activeTab : ''}`}
              onClick={() => setActiveTab('details')}
            >
              Chi tiết đơn hàng
            </button>
            <button
              type="button"
              className={`${styles.tabBtn} ${activeTab === 'tracking' ? styles.activeTab : ''}`}
              onClick={() => setActiveTab('tracking')}
            >
              Theo dõi vận chuyển
            </button>
          </div>

          {activeTab === 'details' && (
            <div className={styles.tabContent}>
              <h2 className={styles.sectionTitle}>Sản phẩm</h2>
              <div className={styles.tableWrapper}>
                <table>
                  <thead>
                    <tr>
                      <th>Tên sản phẩm</th>
                      <th className={styles.textCenter}>Số lượng</th>
                      <th className={styles.textRight}>Thành tiền</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.items.map((item) => (
                      <tr key={item.id}>
                        <td>{item.productName}</td>
                        <td className={styles.textCenter}>{item.quantity}</td>
                        <td className={styles.textRight}>{formatMoney(item.lineTotalCents, data.currency)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'tracking' && (
            <div className={styles.tabContent}>
              {data.deliveryEvents.length === 0 ? (
                <div className={styles.emptyState}>
                  <p className={styles.muted}>Chưa có sự kiện vận chuyển nào.</p>
                </div>
              ) : (
                <ol className={styles.timeline}>
                  {data.deliveryEvents.map((ev) => (
                    <li key={ev.id}>
                      <div className={styles.timelineHeader}>
                        <strong>{ev.statusLabel || ev.statusCode}</strong>
                        <span className={styles.muted}>{formatDate(ev.occurredAt)}</span>
                      </div>
                      <p className={styles.timelineMessage}>{ev.message}</p>
                      {ev.reason ? <p className={styles.timelineReason}>Lý do: {ev.reason}</p> : null}
                    </li>
                  ))}
                </ol>
              )}
            </div>
          )}
        </article>
      )}
    </section>
  )
}
