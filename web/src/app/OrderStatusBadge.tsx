import type { JSX } from 'react'
import styles from './OrderStatusBadge.module.css'

export interface OrderStatusBadgeProps {
  status: string
  label: string
}

export function OrderStatusBadge({ status, label }: OrderStatusBadgeProps): JSX.Element {
  const badgeClass = styles[`badge--${status}`] || ''
  return (
    <span className={`${styles.badge} ${badgeClass}`}>
      {label}
    </span>
  )
}
