import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../auth/authStore'
import { KpiCard } from './KpiCard'
import styles from './dashboard.module.css'

interface ListMeta {
  meta?: { total: number }
  data?: unknown[]
}

async function fetchTotal(
  endpoint: string,
  accessToken: string | null,
): Promise<number> {
  const res = await fetch(`/api/v1/${endpoint}?per_page=1`, {
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
  })
  if (!res.ok) throw new Error(`Failed: ${res.status}`)
  const json = (await res.json()) as ListMeta
  return json.meta?.total ?? (json.data?.length ?? 0)
}

export default function DashboardPage() {
  const accessToken = useAuthStore((s) => s.accessToken)

  const idols = useQuery({
    queryKey: ['kpi', 'idols'],
    queryFn: () => fetchTotal('idols', accessToken),
  })
  const groups = useQuery({
    queryKey: ['kpi', 'groups'],
    queryFn: () => fetchTotal('groups', accessToken),
  })
  const agencies = useQuery({
    queryKey: ['kpi', 'agencies'],
    queryFn: () => fetchTotal('agencies', accessToken),
  })

  return (
    <div className={styles.page}>
      <h1 className={styles.heading}>ダッシュボード</h1>
      <div className={styles.kpiGrid}>
        <KpiCard label="アイドル総数" value={idols.data} isLoading={idols.isLoading} />
        <KpiCard label="グループ総数" value={groups.data} isLoading={groups.isLoading} />
        <KpiCard label="事務所総数" value={agencies.data} isLoading={agencies.isLoading} />
      </div>
    </div>
  )
}
