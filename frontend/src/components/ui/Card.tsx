import styles from './card.module.css'

interface CardProps {
  title?: string
  children: React.ReactNode
  className?: string
}

export function Card({ title, children, className }: CardProps) {
  return (
    <div className={`${styles.card} ${className ?? ''}`}>
      {title && <div className={styles.header}>{title}</div>}
      <div className={styles.body}>{children}</div>
    </div>
  )
}
