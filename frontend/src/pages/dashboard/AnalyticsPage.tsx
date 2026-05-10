import { BarChart2 } from 'lucide-react'
import styles from './dashboard.module.css'

export default function AnalyticsPage() {
  return (
    <div className={styles.page}>
      <h1 className={styles.heading}>分析</h1>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 'var(--space-4)',
        padding: 'var(--space-12)',
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 'var(--radius-md)',
        color: 'var(--color-text-muted)',
      }}>
        <BarChart2 size={40} strokeWidth={1.5} aria-hidden="true" />
        <p style={{ fontSize: 'var(--text-base)', fontWeight: 500 }}>準備中</p>
        <p style={{ fontSize: 'var(--text-sm)' }}>API利用状況の分析データをここに表示します。</p>
      </div>
    </div>
  )
}
