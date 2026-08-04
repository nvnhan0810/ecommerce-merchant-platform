import { useEffect, useState, type FormEvent, type JSX } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import styles from './ProfilePage.module.css'

export function ProfilePage(): JSX.Element {
  const { session, updateProfile, logout } = useAuth()
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (session) {
      setEmail(session.email)
      setDisplayName(session.displayName)
    }
  }, [session])

  if (!session?.isUser) {
    return <Navigate to="/login" replace state={{ from: '/profile' }} />
  }

  async function onSubmit(e: FormEvent): Promise<void> {
    e.preventDefault()
    setError('')
    setMessage('')
    setPending(true)
    try {
      await updateProfile({
        email,
        displayName,
        password: password || undefined,
      })
      setPassword('')
      setMessage('Đã cập nhật thông tin.')
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <h1>Thông tin cá nhân</h1>
        <button type="button" className={styles.logout} onClick={logout}>
          Đăng xuất
        </button>
      </header>
      <form className={styles.form} onSubmit={(e) => void onSubmit(e)}>
        <label>
          Email
          <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
        </label>
        <label>
          Tên hiển thị
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} required />
        </label>
        <label>
          Mật khẩu mới (tuỳ chọn)
          <input
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type="password"
            placeholder="Để trống nếu giữ nguyên"
          />
        </label>
        {error && <p className={styles.error}>{error}</p>}
        {message && <p className={styles.ok}>{message}</p>}
        <button type="submit" disabled={pending}>
          {pending ? 'Đang lưu…' : 'Lưu thay đổi'}
        </button>
      </form>
    </section>
  )
}
