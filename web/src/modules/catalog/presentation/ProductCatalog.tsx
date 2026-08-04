import { useState, type JSX } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ListProductsUseCase } from '../application/list-products'
import { HttpProductRepository } from '../infrastructure/http-product-repository'
import { ListCategoriesUseCase } from '@/modules/categories/application/list-categories'
import { HttpCategoryRepository } from '@/modules/categories/infrastructure/http-category-repository'
import styles from './ProductCatalog.module.css'

const listProducts = new ListProductsUseCase(new HttpProductRepository())
const listCategories = new ListCategoriesUseCase(new HttpCategoryRepository())

type ProductCatalogProps = {
  merchantId?: string
  limit?: number
  /** Hide merchant footer when already browsing a merchant shop page. */
  hideMerchant?: boolean
}

export function ProductCatalog({
  merchantId,
  limit = 40,
  hideMerchant = false,
}: ProductCatalogProps): JSX.Element {
  const [categoryId, setCategoryId] = useState('')
  const { data: categories } = useQuery({
    queryKey: ['catalog', 'categories'],
    queryFn: () => listCategories.execute(),
  })
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['catalog', 'products', merchantId ?? 'all', categoryId || 'all', limit],
    queryFn: () => listProducts.execute(limit, merchantId, categoryId || undefined),
  })

  return (
    <div className={styles.wrap}>
      {categories && categories.length > 0 ? (
        <div className={styles.filters} role="group" aria-label="Lọc danh mục">
          <button
            type="button"
            className={!categoryId ? styles.filterActive : styles.filter}
            onClick={() => setCategoryId('')}
          >
            Tất cả
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              type="button"
              className={categoryId === cat.id ? styles.filterActive : styles.filter}
              onClick={() => setCategoryId(cat.id)}
            >
              {cat.name}
            </button>
          ))}
        </div>
      ) : null}

      {isLoading ? <p className={styles.status}>Đang tải sản phẩm…</p> : null}

      {isError ? (
        <div className={styles.status}>
          <p>Không tải được danh mục: {(error as Error).message}</p>
          <button type="button" onClick={() => void refetch()}>
            Thử lại
          </button>
        </div>
      ) : null}

      {!isLoading && !isError && !data?.length ? (
        <p className={styles.status}>Chưa có sản phẩm.</p>
      ) : null}

      {data && data.length > 0 ? (
        <ul className={styles.grid} aria-label="Danh sách sản phẩm">
          {data.map((product) => (
            <li key={product.id.value} className={styles.card}>
              <Link to={`/products/${product.id.value}`} className={styles.link}>
                {product.imageUrl ? (
                  <img className={styles.thumb} src={product.imageUrl} alt={product.name} />
                ) : (
                  <div className={styles.placeholder} aria-hidden="true" />
                )}
                <h3>{product.name}</h3>
                {product.categories.length > 0 ? (
                  <p className={styles.cats}>
                    {product.categories.map((c) => c.name).join(' · ')}
                  </p>
                ) : null}
                <p className={styles.desc}>{product.description}</p>
                <div className={styles.meta}>
                  <span className={styles.price}>{product.price.format()}</span>
                  <span className={product.isAvailable ? styles.inStock : styles.out}>
                    {product.isAvailable ? `Còn ${product.stock}` : 'Hết hàng'}
                  </span>
                </div>
              </Link>
              {!hideMerchant && product.merchantId ? (
                <Link
                  to={`/merchants/${product.merchantId}`}
                  className={styles.merchant}
                  onClick={(e) => e.stopPropagation()}
                >
                  <span className={styles.merchantLeft}>
                    <span className={styles.merchantAvatar} aria-hidden="true">
                      {product.merchantAvatarUrl ? (
                        <img src={product.merchantAvatarUrl} alt="" />
                      ) : (
                        <span>{(product.merchantDisplayName || '?').charAt(0)}</span>
                      )}
                    </span>
                    <span className={styles.merchantName}>
                      {product.merchantDisplayName || 'Gian hàng'}
                    </span>
                  </span>
                  {product.merchantProvinceName ? (
                    <span className={styles.merchantProvince}>{product.merchantProvinceName}</span>
                  ) : null}
                </Link>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
