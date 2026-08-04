import { useState, type JSX } from 'react'
import { Link, Navigate, useSearchParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { RepayOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import styles from './PaymentResultPage.module.css'

const repayOrder = new RepayOrderUseCase(new HttpOrderRepository())

export function PaymentResultPage(): JSX.Element {
  const { session } = useAuth()
  const [params] = useSearchParams()
  const status = params.get('status') || 'error'
  const paymentId = params.get('payment_id') || ''
  const orderId = params.get('order_id') || ''
  const [repayError, setRepayError] = useState('')

  const repayMutation = useMutation({
    mutationFn: () => repayOrder.execute(orderId),
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
    return <Navigate to="/login" replace state={{ from: `/orders/payment/result?${params.toString()}` }} />
  }

  const paid = status === 'paid'
  const failed = status === 'failed'
  const orderLink = orderId ? `/orders/${orderId}` : '/orders'

  return (
    <section className={styles.page}>
      <div className={`${styles.card} ${paid ? styles.ok : styles.bad}`}>
        <h1>{paid ? 'Thanh toán thành công' : failed ? 'Thanh toán thất bại' : 'Không xác nhận được thanh toán'}</h1>
        <p>
          {paid
            ? 'Giao dịch đã được ghi nhận. Đơn hàng của bạn đang chờ cửa hàng xác nhận.'
            : failed
              ? 'Giao dịch không thành công. Đơn hàng đã được huỷ. Bạn có thể thanh toán lại để mở lại đơn.'
              : 'Không xác minh được phản hồi từ cổng thanh toán. Vui lòng kiểm tra lịch sử đơn hàng.'}
        </p>
        {paymentId ? <p className={styles.meta}>Mã thanh toán: {paymentId}</p> : null}
        {repayError ? <p className={styles.error}>{repayError}</p> : null}
        <div className={styles.actions}>
          {failed && orderId ? (
            <button
              type="button"
              className={styles.primaryBtn}
              disabled={repayMutation.isPending}
              onClick={() => {
                setRepayError('')
                repayMutation.mutate()
              }}
            >
              {repayMutation.isPending ? 'Đang chuyển…' : 'Thanh toán lại'}
            </button>
          ) : null}
          <Link to={orderLink} className={failed && orderId ? styles.secondary : styles.primary}>
            Xem đơn hàng
          </Link>
          <Link to="/" className={styles.secondary}>
            Về trang chủ
          </Link>
        </div>
      </div>
    </section>
  )
}
