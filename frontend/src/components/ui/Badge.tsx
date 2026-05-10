import styles from './badge.module.css'

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'default'

interface BadgeProps {
  variant?: BadgeVariant
  children: React.ReactNode
}

export function Badge({ variant = 'default', children }: BadgeProps) {
  return (
    <span className={`${styles.badge} ${styles[variant]}`}>{children}</span>
  )
}
