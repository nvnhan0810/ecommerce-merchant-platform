import { useState, useEffect, type JSX } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { CreateOrderUseCase } from '@/modules/orders/application/order-use-cases'
import { HttpOrderRepository } from '@/modules/orders/infrastructure/http-order-repository'
import { ListAddressesUseCase } from '@/modules/addresses/application/address-use-cases'
import { HttpAddressRepository } from '@/modules/addresses/infrastructure/http-address-repository'
import { formatFullAddress } from '@/modules/addresses/domain/address'
import { AddressPickerDialog } from './AddressPickerDialog'
import styles from './CheckoutPage.module.css'

const createOrder = new CreateOrderUseCase(new HttpOrderRepository())
const listAddresses = new ListAddressesUseCase(new HttpAddressRepository())

type Step = 1 | 2 | 3

const STEPS: Array<{ id: Step; label: string }> = [
  { id: 1, label: 'Sản phẩm' },
  { id: 2, label: 'Giao hàng' },
  { id: 3, label: 'Xác nhận' },
]

function formatMoney(cents: number, currency: string): string {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: currency || 'VND',
    maximumFractionDigits: 0,
  }).format(cents)
}

export function CheckoutPage(): JSX.Element {
  const { session } = useAuth()
  const { cart, setQuantity, removeItem, clear } = useCart()
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>(1)
  const [note, setNote] = useState('')
  const [shippingName, setShippingName] = useState('')
  const [shippingPhone, setShippingPhone] = useState('')
  const [selectedAddressId, setSelectedAddressId] = useState('')
  const [addressDialogOpen, setAddressDialogOpen] = useState(false)
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  const { data: addresses } = useQuery({
    queryKey: ['user-addresses'],
    queryFn: () => listAddresses.execute(),
    enabled: !!session?.isUser,
  })

  useEffect(() => {
    if (session?.displayName && !shippingName) {
      setShippingName(session.displayName)
    }
  }, [session?.displayName, shippingName])

  useEffect(() => {
    if (addresses && addresses.length > 0 && !selectedAddressId) {
      const defaultAddr = addresses.find((a) => a.isDefault)
      setSelectedAddressId(defaultAddr ? defaultAddr.id : addresses[0].id)
    }
  }, [addresses, selectedAddressId])

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

  const currency = cart.items[0]?.currency || 'VND'
  const selectedAddress = addresses?.find((a) => a.id === selectedAddressId)

  function goNextFromStep1() {
    setError('')
    if (cart.itemCount === 0) {
      setError('Giỏ hàng trống')
      return
    }
    setStep(2)
  }

  function goNextFromStep2() {
    setError('')
    if (!shippingName.trim() || !shippingPhone.trim()) {
      setError('Vui lòng nhập tên người nhận và số điện thoại')
      return
    }
    if (!selectedAddress) {
      setError('Vui lòng chọn địa chỉ giao hàng')
      return
    }
    setStep(3)
  }

  async function placeOrder() {
    setError('')
    if (!selectedAddress) {
      setError('Vui lòng chọn địa chỉ giao hàng')
      setStep(2)
      return
    }

    setPending(true)
    try {
      const orders = await createOrder.execute(
        note,
        cart.items.map((item) => ({ productId: item.productId, quantity: item.quantity })),
        shippingName.trim(),
        shippingPhone.trim(),
        formatFullAddress(selectedAddress),
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
      <header className={styles.pageHeader}>
        <h1>Đặt hàng</h1>
        <p className={styles.meta}>Đăng nhập với {session.email}</p>
      </header>

      <ol className={styles.steps} aria-label="Các bước đặt hàng">
        {STEPS.map((s) => (
          <li
            key={s.id}
            className={`${styles.stepItem} ${step === s.id ? styles.stepActive : ''} ${
              step > s.id ? styles.stepDone : ''
            }`}
          >
            <span className={styles.stepNum}>{s.id}</span>
            <span className={styles.stepLabel}>{s.label}</span>
          </li>
        ))}
      </ol>

      <div className={styles.panel}>
        {step === 1 && (
          <>
            <h2 className={styles.panelTitle}>1. Xác nhận sản phẩm & giá</h2>
            <ul className={styles.productList} aria-label="Sản phẩm đặt hàng">
              {cart.items.map((item) => (
                <li key={item.productId} className={styles.productRow}>
                  <div className={styles.productInfo}>
                    <Link to={`/products/${item.productId}`} className={styles.productName}>
                      {item.name}
                    </Link>
                    <p>{formatMoney(item.unitPriceCents, item.currency)} / sp</p>
                  </div>
                  <div className={styles.productActions}>
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
                    <button
                      type="button"
                      className={styles.removeBtn}
                      onClick={() => removeItem(item.productId)}
                    >
                      Xóa
                    </button>
                  </div>
                </li>
              ))}
            </ul>
            <p className={styles.total}>
              Tổng: <strong>{formatMoney(cart.totalCents, currency)}</strong>
            </p>
          </>
        )}

        {step === 2 && (
          <>
            <h2 className={styles.panelTitle}>2. Thông tin giao hàng</h2>
            <div className={styles.fields}>
              <label>
                Người nhận
                <input
                  value={shippingName}
                  onChange={(e) => setShippingName(e.target.value)}
                  required
                  placeholder="Họ và tên"
                />
              </label>
              <label>
                Số điện thoại
                <input
                  value={shippingPhone}
                  onChange={(e) => setShippingPhone(e.target.value)}
                  required
                  inputMode="tel"
                  placeholder="Số điện thoại nhận hàng"
                />
              </label>

              <div className={styles.addressBlock}>
                <div className={styles.addressBlockHeader}>
                  <strong>Địa chỉ giao hàng</strong>
                  <button
                    type="button"
                    className={styles.chooseAddressBtn}
                    onClick={() => setAddressDialogOpen(true)}
                  >
                    {selectedAddress ? 'Đổi / quản lý' : 'Chọn địa chỉ'}
                  </button>
                </div>
                {selectedAddress ? (
                  <div className={styles.selectedAddress}>
                    <p>
                      <strong>
                        {[selectedAddress.wardName, selectedAddress.provinceName]
                          .filter(Boolean)
                          .join(', ') || 'Địa chỉ'}
                      </strong>
                      {selectedAddress.isDefault ? ' · Mặc định' : ''}
                    </p>
                    <p>{formatFullAddress(selectedAddress)}</p>
                  </div>
                ) : (
                  <p className={styles.addressEmpty}>Chưa chọn địa chỉ giao hàng.</p>
                )}
              </div>

              <label>
                Ghi chú đơn hàng
                <textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} />
              </label>
            </div>
          </>
        )}

        {step === 3 && (
          <>
            <h2 className={styles.panelTitle}>3. Kiểm tra lại đơn hàng</h2>

            <section className={styles.reviewSection}>
              <div className={styles.reviewHead}>
                <h3>Sản phẩm</h3>
                <button type="button" className={styles.linkBtn} onClick={() => setStep(1)}>
                  Sửa
                </button>
              </div>
              <ul className={styles.reviewList}>
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
                Tổng: <strong>{formatMoney(cart.totalCents, currency)}</strong>
              </p>
            </section>

            <section className={styles.reviewSection}>
              <div className={styles.reviewHead}>
                <h3>Giao hàng</h3>
                <button type="button" className={styles.linkBtn} onClick={() => setStep(2)}>
                  Sửa
                </button>
              </div>
              <dl className={styles.reviewDl}>
                <div>
                  <dt>Người nhận</dt>
                  <dd>{shippingName.trim()}</dd>
                </div>
                <div>
                  <dt>Số điện thoại</dt>
                  <dd>{shippingPhone.trim()}</dd>
                </div>
                <div>
                  <dt>Địa chỉ</dt>
                  <dd>{selectedAddress ? formatFullAddress(selectedAddress) : '—'}</dd>
                </div>
                {note.trim() && (
                  <div>
                    <dt>Ghi chú</dt>
                    <dd>{note.trim()}</dd>
                  </div>
                )}
              </dl>
            </section>
          </>
        )}

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.nav}>
          {step > 1 ? (
            <button type="button" className={styles.secondaryBtn} onClick={() => setStep((step - 1) as Step)}>
              Quay lại
            </button>
          ) : (
            <Link to="/cart" className={styles.secondaryLink}>
              Về giỏ hàng
            </Link>
          )}

          {step === 1 && (
            <button type="button" className={styles.primaryBtn} onClick={goNextFromStep1}>
              Tiếp tục
            </button>
          )}
          {step === 2 && (
            <button type="button" className={styles.primaryBtn} onClick={goNextFromStep2}>
              Tiếp tục
            </button>
          )}
          {step === 3 && (
            <button type="button" className={styles.primaryBtn} disabled={pending} onClick={() => void placeOrder()}>
              {pending ? 'Đang đặt…' : 'Xác nhận đặt hàng'}
            </button>
          )}
        </div>
      </div>

      <AddressPickerDialog
        open={addressDialogOpen}
        selectedId={selectedAddressId}
        onClose={() => setAddressDialogOpen(false)}
        onSelect={setSelectedAddressId}
      />
    </section>
  )
}
