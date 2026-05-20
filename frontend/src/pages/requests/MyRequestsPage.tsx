import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { ClipboardList, Trash2, Clock, CheckCircle2, XCircle, AlertCircle } from 'lucide-react'
import { useAuth } from '../../auth/useAuth'
import { Skeleton } from '../../components/ui/Skeleton'
import styles from './request.module.css'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

interface Submission {
  id: string
  target_type: string
  status: string
  created_at: string
}

interface RemovalRequest {
  id: string
  target_type: string
  target_id: string
  status: string
  created_at: string
}

async function authedGet<T>(path: string, accessToken: string, idToken: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'X-ID-Token': idToken,
    },
  })
  if (!res.ok) throw new Error(`request failed: ${res.status}`)
  return (await res.json()) as T
}

const TARGET_TYPE_LABEL: Record<string, string> = {
  idol: 'アイドル',
  group: 'グループ',
  agency: '事務所',
  event: 'イベント',
}

const STATUS_CONFIG: Record<string, { label: string; icon: typeof Clock; color: string }> = {
  pending: { label: '審査中', icon: Clock, color: 'var(--color-warning)' },
  approved: { label: '承認', icon: CheckCircle2, color: 'var(--color-success)' },
  rejected: { label: '却下', icon: XCircle, color: 'var(--color-danger)' },
}

function StatusBadge({ status }: { status: string }) {
  const config = STATUS_CONFIG[status] ?? { label: status, icon: AlertCircle, color: 'var(--color-text-muted)' }
  const Icon = config.icon
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', fontSize: 'var(--text-xs)', fontWeight: 700, color: config.color }}>
      <Icon size={12} aria-hidden="true" />
      {config.label}
    </span>
  )
}

function SubmissionItem({ item }: { item: Submission }) {
  return (
    <div className={styles.item}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-3)' }}>
        <strong style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text)' }}>
          {TARGET_TYPE_LABEL[item.target_type] ?? item.target_type}
        </strong>
        <StatusBadge status={item.status} />
      </div>
      <span className={styles.meta}>ID: {item.id}</span>
      <span className={styles.meta}>{format(new Date(item.created_at), 'yyyy/MM/dd HH:mm')}</span>
    </div>
  )
}

function RemovalItem({ item }: { item: RemovalRequest }) {
  return (
    <div className={styles.item}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-3)' }}>
        <strong style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text)' }}>
          {TARGET_TYPE_LABEL[item.target_type] ?? item.target_type}
        </strong>
        <StatusBadge status={item.status} />
      </div>
      <span className={styles.meta}>対象ID: {item.target_id}</span>
      <span className={styles.meta}>{format(new Date(item.created_at), 'yyyy/MM/dd HH:mm')}</span>
    </div>
  )
}

export default function MyRequestsPage() {
  const { accessToken, idToken, isLoggedIn } = useAuth()

  const submissions = useQuery({
    queryKey: ['me', 'submissions'],
    enabled: isLoggedIn && !!accessToken && !!idToken,
    queryFn: () =>
      authedGet<{ submissions: Submission[] }>('/me/submissions', accessToken!, idToken!),
  })
  const removals = useQuery({
    queryKey: ['me', 'removal-requests'],
    enabled: isLoggedIn && !!accessToken && !!idToken,
    queryFn: () =>
      authedGet<{ removal_requests: RemovalRequest[] }>('/me/removal-requests', accessToken!, idToken!),
  })

  if (!isLoggedIn) {
    return (
      <div className={styles.page}>
        <div className={styles.toolbar}>
          <h1 className={styles.title}>申請履歴</h1>
        </div>
        <div className={styles.panel} style={{ textAlign: 'center', padding: 'var(--space-12)' }}>
          <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--text-sm)' }}>
            申請履歴を表示するにはサインインしてください。
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>申請履歴</h1>
      </div>
      <div className={styles.grid}>
        <section className={styles.panel}>
          <h2 style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: 'var(--text-base)', fontWeight: 700, color: 'var(--color-text)', marginBottom: 'var(--space-4)' }}>
            <ClipboardList size={16} aria-hidden="true" />
            登録申請
          </h2>
          <div className={styles.list}>
            {submissions.isLoading && (
              <>
                <Skeleton height="72px" />
                <Skeleton height="72px" />
              </>
            )}
            {!submissions.isLoading && (submissions.data?.submissions ?? []).length === 0 && (
              <span className={styles.meta}>申請はありません。</span>
            )}
            {(submissions.data?.submissions ?? []).map((item) => (
              <SubmissionItem key={item.id} item={item} />
            ))}
          </div>
        </section>

        <section className={styles.panel}>
          <h2 style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: 'var(--text-base)', fontWeight: 700, color: 'var(--color-text)', marginBottom: 'var(--space-4)' }}>
            <Trash2 size={16} aria-hidden="true" />
            削除申請
          </h2>
          <div className={styles.list}>
            {removals.isLoading && (
              <>
                <Skeleton height="72px" />
                <Skeleton height="72px" />
              </>
            )}
            {!removals.isLoading && (removals.data?.removal_requests ?? []).length === 0 && (
              <span className={styles.meta}>申請はありません。</span>
            )}
            {(removals.data?.removal_requests ?? []).map((item) => (
              <RemovalItem key={item.id} item={item} />
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}
