import type { JSX } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/modules/auth/presentation/AuthProvider'
import { AddressList } from './AddressList'
import styles from './AddressesPage.module.css'

export function AddressesPage(): JSX.Element {
  const { session } = useAuth()

  if (!session?.isUser) {
    return <Navigate to="/login" replace state={{ from: '/addresses' }} />
  }

  return (
    <section className={styles.page}>
      <AddressList />
    </section>
  )
}
