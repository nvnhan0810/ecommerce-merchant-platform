import type { JSX } from 'react'
import { Link, Outlet } from 'react-router-dom'
import styles from './AppShell.module.css'

export function AppShell(): JSX.Element {
  return (
    <div className={styles.shell}>
      <header className={styles.header}>
        <Link to="/" className={styles.brand}>
          Ecomerce
        </Link>
        <nav className={styles.nav} aria-label="Chính">
          <Link to="/">Trang chủ</Link>
          <Link to="/cart">Giỏ hàng</Link>
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
