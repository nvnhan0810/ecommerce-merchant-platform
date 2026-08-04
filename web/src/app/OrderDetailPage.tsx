import { useState, type JSX } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { GetMyOrderUseCase, RepayOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import { OrderStatusBadge } from './OrderStatusBadge'
import styles from './OrderDetailPage.module.css'

const orderRepo = new HttpOrderRepository()
const getOrder = new GetMyOrderUseCase(orderRepo)
const repayOrder = new RepayOrderUseCase(orderRepo)

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
  const [repayError, setRepayError] = useState('')
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['user', 'orders', id],
    queryFn: () => getOrder.execute(id),
    enabled: Boolean(session?.isUser && id),
  })

  const repayMutation = useMutation({
    mutationFn: () => repayOrder.execute(id),
    onSuccess: (result) => {
      const url = result.payment?.paymentUrl
      if (url) {
        window.location.assign(url)
        return
      }
      setRepayError('Không tạo được link thanh toán.')
    },
    onError: (err) => {
      setRepayError((err as Error).message || 'Thanh toán lại thất bại.')
    },
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

          {data.canRepay ? (
            <div className={styles.repayBox}>
              <p>
                {data.paymentStatus === 'failed' || data.status === 'cancelled'
                  ? 'Thanh toán không thành công. Bạn có thể thanh toán lại để tiếp tục đơn hàng.'
                  : 'Đơn hàng đang chờ thanh toán. Tiếp tục thanh toán để hoàn tất.'}
              </p>
              {repayError ? <p className={styles.repayError}>{repayError}</p> : null}
              <button
                type="button"
                className={styles.repayBtn}
                disabled={repayMutation.isPending}
                onClick={() => {
                  setRepayError('')
                  repayMutation.mutate()
                }}
              >
                {repayMutation.isPending ? 'Đang chuyển…' : 'Thanh toán lại'}
              </button>
            </div>
          ) : null}

          <div className={styles.infoGrid}>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Tổng tiền</span>
              <strong className={styles.totalPrice}>{formatMoney(data.totalCents, data.currency)}</strong>
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>Thanh toán</span>
              <span>
                {data.paymentMethodLabel} · {data.paymentStatusLabel}
              </span>
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

          <div className={styles.addressBox}>
            <h3 className={styles.sectionTitle}>Thông tin giao hàng</h3>
            <dl className={styles.shippingDl}>
              <div className={styles.shippingRow}>
                <div>
                  <dt>Người nhận</dt>
                  <dd>{data.shippingName || '—'}</dd>
                </div>
                <div>
                  <dt>Số điện thoại</dt>
                  <dd>{data.shippingPhone || '—'}</dd>
                </div>
              </div>
              <div>
                <dt>Địa chỉ</dt>
                <dd>{data.shippingAddress || '—'}</dd>
              </div>
            </dl>
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
                        <td className={styles.textRight}>
                          {formatMoney(item.lineTotalCents, data.currency)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'tracking' && (
            <div className={styles.tabContent}>
              <section className={styles.carrierSection}>
                <h2 className={styles.sectionTitle}>Thông tin vận chuyển</h2>
                <dl className={styles.shippingDl}>
                  <div>
                    <dt>Mã vận đơn</dt>
                    <dd className={styles.codeInline}>{data.deliveryTrackingCode || '—'}</dd>
                  </div>
                  <div>
                    <dt>Đơn vị vận chuyển</dt>
                    <dd>{data.deliveryCarrier || 'internal'}</dd>
                  </div>
                </dl>
              </section>

              <section>
                <h2 className={styles.sectionTitle}>Lịch trình</h2>
                {data.deliveryEvents.length === 0 ? (
                  <p className={styles.muted}>Chưa có sự kiện vận chuyển nào.</p>
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
              </section>
            </div>
          )}
        </article>
      )}
    </section>
  )
}
