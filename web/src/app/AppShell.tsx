import type { JSX } from 'react'
import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import styles from './AppShell.module.css'

export function AppShell(): JSX.Element {
  const { session, logout } = useAuth()
  const { cart } = useCart()

  return (
    <div className={styles.shell}>
      <header className={styles.header}>
        <Link to="/" className={styles.brand}>
          Ecomerce
        </Link>
        <nav className={styles.nav} aria-label="Chính">
          <Link to="/">Sản phẩm</Link>
          <Link to="/cart">Giỏ hàng{cart.itemCount > 0 ? ` (${cart.itemCount})` : ''}</Link>
          {session?.isUser ? (
            <>
              <Link to="/orders">Đơn hàng</Link>
              <Link to="/profile">Tài khoản</Link>
              <button type="button" className={styles.linkBtn} onClick={logout}>
                Đăng xuất
              </button>
            </>
          ) : (
            <Link to="/login">Đăng nhập</Link>
          )}
        </nav>
      </header>
      <main className={styles.main}>
        <Outlet />
      </main>
      <footer className={styles.footer}>
        <span>ecomerce.nvnhan0810.com</span>
      </footer>
    </div>
  )
}
