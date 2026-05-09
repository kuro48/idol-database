import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'
import { format } from 'date-fns'
import { Search } from 'lucide-react'
import { DataTable } from '../../components/table/DataTable'
import { Pagination } from '../../components/table/Pagination'
import styles from '../idols/idol-list.module.css'

interface Event {
  id: string
  name: string
  date?: string
  created_at?: string
}

interface EventsResponse {
  data: Event[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchEvents(page: number, perPage: number, q: string): Promise<EventsResponse> {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage), ...(q ? { q } : {}) })
  const res = await fetch(`/api/v1/events?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch events: ${res.status}`)
  return res.json() as Promise<EventsResponse>
}

const COLUMNS: ColumnDef<Event, unknown>[] = [
  { accessorKey: 'name', header: '名前' },
  {
    accessorKey: 'date',
    header: '開催日',
    cell: ({ getValue }) => {
      const v = getValue() as string | undefined
      return v ? format(new Date(v), 'yyyy/MM/dd') : '—'
    },
  },
  {
    accessorKey: 'created_at',
    header: '登録日',
    cell: ({ getValue }) => {
      const v = getValue() as string | undefined
      return v ? format(new Date(v), 'yyyy/MM/dd') : '—'
    },
  },
]

export default function EventListPage() {
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)
  const [search, setSearch] = useState('')
  const [q, setQ] = useState('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['events', page, perPage, q],
    queryFn: () => fetchEvents(page, perPage, q),
  })

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    setQ(search)
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>イベント</h1>
        <form onSubmit={handleSearch} className={styles.searchForm}>
          <div className={styles.searchWrapper}>
            <Search size={14} className={styles.searchIcon} aria-hidden="true" />
            <input type="search" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="イベントを検索…" className={styles.searchInput} />
          </div>
        </form>
      </div>
      {isError && <div className={styles.error} role="alert">イベントの読み込みに失敗しました。もう一度お試しください。</div>}
      <div className={styles.tableCard}>
        <DataTable columns={COLUMNS} data={data?.data ?? []} isLoading={isLoading} emptyMessage="イベントが見つかりません。" />
        {data?.meta && <Pagination page={page} perPage={perPage} total={data.meta.total} onPageChange={setPage} onPerPageChange={(n) => { setPerPage(n); setPage(1) }} />}
      </div>
    </div>
  )
}
