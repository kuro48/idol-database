import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'
import { format } from 'date-fns'
import { Search } from 'lucide-react'
import { DataTable } from '../../components/table/DataTable'
import { Pagination } from '../../components/table/Pagination'
import styles from './idol-list.module.css'

interface Idol {
  id: string
  name: string
  agency?: string
  created_at?: string
}

interface IdolsResponse {
  data: Idol[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchIdols(page: number, perPage: number, q: string): Promise<IdolsResponse> {
  const params = new URLSearchParams({
    page: String(page),
    per_page: String(perPage),
    ...(q ? { q } : {}),
  })
  const res = await fetch(`/api/v1/idols?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch idols: ${res.status}`)
  return res.json() as Promise<IdolsResponse>
}

const COLUMNS: ColumnDef<Idol, unknown>[] = [
  { accessorKey: 'name', header: '名前' },
  {
    accessorKey: 'agency',
    header: '事務所',
    cell: ({ getValue }) => (getValue() as string) ?? '—',
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

export default function IdolListPage() {
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)
  const [search, setSearch] = useState('')
  const [q, setQ] = useState('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['idols', page, perPage, q],
    queryFn: () => fetchIdols(page, perPage, q),
  })

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    setQ(search)
  }

  function handlePerPageChange(newPerPage: number) {
    setPerPage(newPerPage)
    setPage(1)
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>アイドル</h1>
        <form onSubmit={handleSearch} className={styles.searchForm}>
          <div className={styles.searchWrapper}>
            <Search size={14} className={styles.searchIcon} aria-hidden="true" />
            <input
              type="search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="アイドルを検索…"
              className={styles.searchInput}
            />
          </div>
        </form>
      </div>

      {isError && (
        <div className={styles.error} role="alert">
          アイドルの読み込みに失敗しました。もう一度お試しください。
        </div>
      )}

      <div className={styles.tableCard}>
        <DataTable
          columns={COLUMNS}
          data={data?.data ?? []}
          isLoading={isLoading}
          emptyMessage="アイドルが見つかりません。"
        />
        {data?.meta && (
          <Pagination
            page={page}
            perPage={perPage}
            total={data.meta.total}
            onPageChange={setPage}
            onPerPageChange={handlePerPageChange}
          />
        )}
      </div>
    </div>
  )
}
