import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'
import { format } from 'date-fns'
import { DataTable } from '../../components/table/DataTable'
import { Pagination } from '../../components/table/Pagination'
import styles from '../idols/idol-list.module.css'

interface Release {
  id: string
  title: string
  release_date?: string
  type?: string
}

interface ReleasesResponse {
  data: Release[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchReleases(page: number, perPage: number): Promise<ReleasesResponse> {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) })
  const res = await fetch(`/api/v1/releases?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch releases: ${res.status}`)
  return res.json() as Promise<ReleasesResponse>
}

const COLUMNS: ColumnDef<Release, unknown>[] = [
  { accessorKey: 'title', header: 'タイトル' },
  { accessorKey: 'type', header: '種別', cell: ({ getValue }) => (getValue() as string) ?? '—' },
  {
    accessorKey: 'release_date',
    header: '発売日',
    cell: ({ getValue }) => {
      const v = getValue() as string | undefined
      return v ? format(new Date(v), 'yyyy/MM/dd') : '—'
    },
  },
]

export default function ReleaseListPage() {
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)

  const { data, isLoading, isError } = useQuery({
    queryKey: ['releases', page, perPage],
    queryFn: () => fetchReleases(page, perPage),
  })

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>リリース</h1>
      </div>
      {isError && <div className={styles.error} role="alert">リリースの読み込みに失敗しました。もう一度お試しください。</div>}
      <div className={styles.tableCard}>
        <DataTable columns={COLUMNS} data={data?.data ?? []} isLoading={isLoading} emptyMessage="リリースが見つかりません。" />
        {data?.meta && <Pagination page={page} perPage={perPage} total={data.meta.total} onPageChange={setPage} onPerPageChange={(n) => { setPerPage(n); setPage(1) }} />}
      </div>
    </div>
  )
}
