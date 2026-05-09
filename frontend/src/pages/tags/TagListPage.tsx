import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'
import { DataTable } from '../../components/table/DataTable'
import { Pagination } from '../../components/table/Pagination'
import styles from '../idols/idol-list.module.css'

interface Tag {
  id: string
  name: string
  count?: number
}

interface TagsResponse {
  data: Tag[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchTags(page: number, perPage: number): Promise<TagsResponse> {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) })
  const res = await fetch(`/api/v1/tags?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch tags: ${res.status}`)
  return res.json() as Promise<TagsResponse>
}

const COLUMNS: ColumnDef<Tag, unknown>[] = [
  { accessorKey: 'name', header: 'タグ名' },
  { accessorKey: 'count', header: '件数', cell: ({ getValue }) => (getValue() as number | undefined)?.toLocaleString('ja-JP') ?? '—' },
]

export default function TagListPage() {
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)

  const { data, isLoading, isError } = useQuery({
    queryKey: ['tags', page, perPage],
    queryFn: () => fetchTags(page, perPage),
  })

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>タグ</h1>
      </div>
      {isError && <div className={styles.error} role="alert">タグの読み込みに失敗しました。もう一度お試しください。</div>}
      <div className={styles.tableCard}>
        <DataTable columns={COLUMNS} data={data?.data ?? []} isLoading={isLoading} emptyMessage="タグが見つかりません。" />
        {data?.meta && <Pagination page={page} perPage={perPage} total={data.meta.total} onPageChange={setPage} onPerPageChange={(n) => { setPerPage(n); setPage(1) }} />}
      </div>
    </div>
  )
}
