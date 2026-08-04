import { useState, useRef, useEffect, type JSX } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { ShoppingCart, User, LogOut, Package, ChevronDown } from 'lucide-react'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { useCart } from '@/modules/cart/presentation/CartProvider'
import styles from './AppShell.module.css'

export function AppShell(): JSX.Element {
  const { session, logout } = useAuth()
  const { cart } = useCart()
  const location = useLocation()
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setIsUserMenuOpen(false)
  }, [location.pathname])

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div className={styles.shell}>
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <Link to="/" className={styles.brand}>
            Ecomerce
          </Link>
          <nav className={styles.mainNav} aria-label="Chính">
            <Link to="/">Sản phẩm</Link>
          </nav>
        </div>
        
        <div className={styles.headerRight}>
          <Link to="/cart" className={styles.cartBtn} aria-label="Giỏ hàng">
            <ShoppingCart size={20} />
            {cart.itemCount > 0 && (
              <span className={styles.cartBadge}>{cart.itemCount}</span>
            )}
          </Link>
          
          {session?.isUser ? (
            <div className={styles.userMenuWrapper} ref={menuRef}>
              <button 
                type="button" 
                className={styles.userMenuBtn}
                onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                aria-expanded={isUserMenuOpen}
              >
                <div className={styles.avatar}>
                  <User size={18} />
                </div>
                <span className={styles.userName}>{session.displayName}</span>
                <ChevronDown size={16} className={styles.chevron} />
              </button>
              
              {isUserMenuOpen && (
                <div className={styles.dropdownMenu}>
                  <Link to="/profile" className={styles.dropdownItem}>
                    <User size={16} />
                    <span>Tài khoản</span>
                  </Link>
                  <Link to="/orders" className={styles.dropdownItem}>
                    <Package size={16} />
                    <span>Đơn hàng của tôi</span>
                  </Link>
                  <div className={styles.dropdownDivider} />
                  <button type="button" className={styles.dropdownItem} onClick={logout}>
                    <LogOut size={16} />
                    <span>Đăng xuất</span>
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link to="/login" className={styles.loginBtn}>Đăng nhập</Link>
          )}
        </div>
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
