import type { JSX } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { GetMerchantUseCase } from '@/modules/merchants/application/merchant-use-cases'
import { HttpMerchantRepository } from '@/modules/merchants/infrastructure/http-merchant-repository'
import { ProductCatalog } from '@/modules/catalog/presentation/ProductCatalog'
import styles from './MerchantPage.module.css'

const getMerchant = new GetMerchantUseCase(new HttpMerchantRepository())

export function MerchantPage(): JSX.Element {
  const { id = '' } = useParams()
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['merchants', id],
    queryFn: () => getMerchant.execute(id),
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
  if (!data) return <p>Không tìm thấy gian hàng.</p>

  return (
    <section className={styles.page}>
      <Link className={styles.back} to="/">
        ← Danh sách sản phẩm
      </Link>
      <header className={styles.hero}>
        <div className={styles.identity}>
          <div className={styles.avatar} aria-hidden="true">
            {data.avatarUrl ? (
              <img src={data.avatarUrl} alt="" />
            ) : (
              <span>{data.displayName.charAt(0)}</span>
            )}
          </div>
          <div>
            <p className={styles.kicker}>Gian hàng</p>
            <h1>{data.displayName}</h1>
            {data.formattedLocation ? (
              <p className={styles.address}>{data.formattedLocation}</p>
            ) : (
              <p className={styles.lead}>Sản phẩm đang bán từ gian hàng này.</p>
            )}
          </div>
        </div>
      </header>
      <ProductCatalog merchantId={data.id} hideMerchant />
    </section>
  )
}
