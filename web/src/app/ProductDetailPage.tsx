import type { JSX } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { GetProductUseCase } from '@/modules/catalog/application/list-products'
import { HttpProductRepository } from '@/modules/catalog/infrastructure/http-product-repository'
import { CartItem } from '@/modules/cart/domain/cart'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import styles from './ProductDetailPage.module.css'

const getProduct = new GetProductUseCase(new HttpProductRepository())

export function ProductDetailPage(): JSX.Element {
  const { id = '' } = useParams()
  const { addItem } = useCart()
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['catalog', 'products', id],
    queryFn: () => getProduct.execute(id),
    enabled: Boolean(id),
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
        1,
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
          <button type="button" disabled={!data.isAvailable} onClick={onAdd}>
            Thêm vào giỏ
          </button>
        </div>
      </div>
    </article>
  )
}
