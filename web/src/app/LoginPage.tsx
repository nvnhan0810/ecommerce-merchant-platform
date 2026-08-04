import { useState, type FormEvent, type JSX } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import styles from './LoginPage.module.css'

export function LoginPage(): JSX.Element {
  const { session, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from || '/'
  const [email, setEmail] = useState('buyer@ecomerce.local')
  const [password, setPassword] = useState('Buyer@123456')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  if (session?.isUser) {
    return <Navigate to={from} replace />
  }

  async function onSubmit(e: FormEvent): Promise<void> {
    e.preventDefault()
    setError('')
    setPending(true)
    try {
      await login(email, password)
      void navigate(from, { replace: true })
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <section className={styles.page}>
      <h1>Đăng nhập</h1>
      <p className={styles.lead}>Đăng nhập để đặt hàng và xem lịch sử đơn.</p>
      <form className={styles.form} onSubmit={(e) => void onSubmit(e)}>
        <label>
          Email
          <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
        </label>
        <label>
          Mật khẩu
          <input
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type="password"
            required
          />
        </label>
        {error && <p className={styles.error}>{error}</p>}
        <button type="submit" disabled={pending}>
          {pending ? 'Đang đăng nhập…' : 'Đăng nhập'}
        </button>
      </form>
    </section>
  )
}
