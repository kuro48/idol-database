import { Skeleton } from '../../components/ui/Skeleton'
import styles from './dashboard.module.css'

interface KpiCardProps {
  label: string
  value: number | undefined
  isLoading?: boolean
}

export function KpiCard({ label, value, isLoading = false }: KpiCardProps) {
  return (
    <div className={styles.kpiCard}>
      <div className={styles.kpiAccentBar} aria-hidden="true" />
      <div className={styles.kpiBody}>
        {isLoading ? (
          <Skeleton width="80px" height="2rem" />
        ) : (
          <span className={styles.kpiValue}>
            {value?.toLocaleString() ?? '—'}
          </span>
        )}
        <span className={styles.kpiLabel}>{label}</span>
      </div>
    </div>
  )
}
