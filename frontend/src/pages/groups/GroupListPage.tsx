import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { Search, Music, Building2, CalendarDays } from 'lucide-react'
import { Skeleton } from '../../components/ui/Skeleton'
import { Pagination } from '../../components/table/Pagination'
import styles from '../idols/idol-list.module.css'

interface Group {
  id: string
  name: string
  agency?: string
  created_at?: string
}

interface GroupsResponse {
  data: Group[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchGroups(page: number, perPage: number, q: string): Promise<GroupsResponse> {
  const params = new URLSearchParams({
    page: String(page),
    per_page: String(perPage),
    ...(q ? { q } : {}),
  })
  const res = await fetch(`/api/v1/groups?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch groups: ${res.status}`)
  return res.json() as Promise<GroupsResponse>
}

const SKELETON_COUNT = 12

export default function GroupListPage() {
  const [page, setPage] = useState(1)
  const [perPage] = useState(24)
  const [search, setSearch] = useState('')
  const [q, setQ] = useState('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['groups', page, perPage, q],
    queryFn: () => fetchGroups(page, perPage, q),
  })

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setPage(1)
    setQ(search)
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <div>
          <h1 className={styles.title}>グループ</h1>
          {data?.meta && (
            <p className={styles.count}>{data.meta.total.toLocaleString('ja-JP')} 件</p>
          )}
        </div>
        <form onSubmit={handleSearch} className={styles.searchForm}>
          <div className={styles.searchWrapper}>
            <Search size={14} className={styles.searchIcon} aria-hidden="true" />
            <input
              type="search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="グループを検索…"
              className={styles.searchInput}
            />
          </div>
        </form>
      </div>

      {isError && (
        <div className={styles.error} role="alert">
          グループの読み込みに失敗しました。もう一度お試しください。
        </div>
      )}

      <div className={styles.grid}>
        {isLoading
          ? Array.from({ length: SKELETON_COUNT }).map((_, i) => (
              <div className={styles.skeletonCard} key={i} aria-hidden="true">
                <Skeleton width="44px" height="44px" />
                <Skeleton width="70%" height="1.2rem" />
                <Skeleton width="50%" height="0.9rem" />
              </div>
            ))
          : (data?.data ?? []).map((group) => (
              <article className={styles.card} key={group.id}>
                <div className={styles.cardIcon} aria-hidden="true">
                  <Music size={20} />
                </div>
                <p className={styles.cardName}>{group.name}</p>
                <div className={styles.cardMeta}>
                  {group.agency && (
                    <span className={styles.cardMetaItem}>
                      <Building2 size={12} aria-hidden="true" />
                      {group.agency}
                    </span>
                  )}
                  {group.created_at && (
                    <span className={styles.cardMetaItem}>
                      <CalendarDays size={12} aria-hidden="true" />
                      {format(new Date(group.created_at), 'yyyy/MM/dd')}
                    </span>
                  )}
                </div>
              </article>
            ))}
      </div>

      {data?.meta && (
        <Pagination
          page={page}
          perPage={perPage}
          total={data.meta.total}
          onPageChange={setPage}
          onPerPageChange={() => {}}
        />
      )}
    </div>
  )
}
