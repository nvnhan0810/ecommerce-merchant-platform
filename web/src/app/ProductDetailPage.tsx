import { useState, type JSX } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { GetProductUseCase } from '@/modules/catalog/application/list-products'
import { HttpProductRepository } from '@/modules/catalog/infrastructure/http-product-repository'
import { GetMerchantUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'
import { CartItem } from '@/modules/cart/domain/cart'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import styles from './ProductDetailPage.module.css'

const getProduct = new GetProductUseCase(new HttpProductRepository())
const getMerchant = new GetMerchantUseCase(new HttpMerchantRepository())

export function ProductDetailPage(): JSX.Element {
  const { id = '' } = useParams()
  const { addItem } = useCart()
  const [quantity, setQuantity] = useState(1)
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['catalog', 'products', id],
    queryFn: () => getProduct.execute(id),
    enabled: Boolean(id),
  })
  const merchantId = data?.merchantId ?? ''
  const {
    data: merchant,
    isLoading: merchantLoading,
  } = useQuery({
    queryKey: ['merchants', merchantId],
    queryFn: () => getMerchant.execute(merchantId),
    enabled: Boolean(merchantId),
  })

  if (isLoading) return <p>Đang tải…</p>
  if (isError) {
    return (
      <div>
        <p>{(error as Error).message}</p>
        <button type="button" onClick={() => void refetch()}>
          Thử lại
        </button>
      </div>
    )
  }
  if (!data) return <p>Không tìm thấy sản phẩm.</p>

  function onAdd(): void {
    if (!data?.isAvailable) return
    addItem(
      new CartItem(
        data.id.value,
        data.merchantId,
        data.name,
        data.price.amountCents,
        data.price.currency,
        quantity,
        data.imageUrl,
      ),
    )
  }

  return (
    <article className={styles.page}>
      <Link className={styles.back} to="/">
        ← Danh sách sản phẩm
      </Link>
      <div className={styles.layout}>
        {data.imageUrl ? (
          <img className={styles.hero} src={data.imageUrl} alt={data.name} />
        ) : (
          <div className={styles.heroPlaceholder}>Chưa có ảnh</div>
        )}
        <div className={styles.info}>
          <h1>{data.name}</h1>
          <p className={styles.price}>{data.price.format()}</p>
          <p className={styles.stock}>
            {data.isAvailable ? `Còn ${data.stock} sản phẩm` : 'Hết hàng'}
          </p>
          <p className={styles.desc}>{data.description || '—'}</p>
          <div className={styles.actions}>
            <input
              type="number"
              min="1"
              max={data.stock}
              value={quantity}
              onChange={(e) => setQuantity(Math.max(1, Math.min(data.stock, Number(e.target.value))))}
              disabled={!data.isAvailable}
              className={styles.quantityInput}
            />
            <button type="button" disabled={!data.isAvailable} onClick={onAdd}>
              Thêm vào giỏ
            </button>
          </div>
        </div>
      </div>

      {merchantLoading ? (
        <section className={styles.shopSection} aria-busy="true">
          <p className={styles.merchant}>Đang tải gian hàng…</p>
        </section>
      ) : merchant ? (
        <section className={styles.shopSection} aria-label="Thông tin gian hàng">
          <div className={styles.merchantCard}>
            <div className={styles.merchantAvatar} aria-hidden="true">
              {merchant.avatarUrl ? (
                <img src={merchant.avatarUrl} alt="" />
              ) : (
                <span>{merchant.displayName.charAt(0)}</span>
              )}
            </div>
            <div className={styles.merchantInfo}>
              <p className={styles.merchantName}>{merchant.displayName}</p>
              {merchant.wardName || merchant.provinceName ? (
                <p className={styles.merchantAddress}>
                  {[merchant.wardName, merchant.provinceName].filter(Boolean).join(', ')}
                </p>
              ) : null}
            </div>
            <Link className={styles.viewShop} to={`/merchants/${merchant.id}`}>
              Xem
            </Link>
          </div>
        </section>
      ) : null}
    </article>
  )
}
