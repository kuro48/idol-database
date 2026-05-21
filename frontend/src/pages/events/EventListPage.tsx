import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { Search, Calendar, CalendarDays } from 'lucide-react'
import { Skeleton } from '../../components/ui/Skeleton'
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
  const params = new URLSearchParams({
    page: String(page),
    per_page: String(perPage),
    ...(q ? { q } : {}),
  })
  const res = await fetch(`/api/v1/events?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch events: ${res.status}`)
  return res.json() as Promise<EventsResponse>
}

const SKELETON_COUNT = 12

export default function EventListPage() {
  const [page, setPage] = useState(1)
  const [perPage] = useState(24)
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
        <div>
          <h1 className={styles.title}>イベント</h1>
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
              placeholder="イベントを検索…"
              className={styles.searchInput}
            />
          </div>
        </form>
      </div>

      {isError && (
        <div className={styles.error} role="alert">
          イベントの読み込みに失敗しました。もう一度お試しください。
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
          : (data?.data ?? []).map((event) => (
              <article className={styles.card} key={event.id}>
                <div className={styles.cardIcon} aria-hidden="true">
                  <Calendar size={20} />
                </div>
                <p className={styles.cardName}>{event.name}</p>
                <div className={styles.cardMeta}>
                  {event.date && (
                    <span className={styles.badge}>
                      {format(new Date(event.date), 'yyyy/MM/dd')}
                    </span>
                  )}
                  {event.created_at && (
                    <span className={styles.cardMetaItem}>
                      <CalendarDays size={12} aria-hidden="true" />
                      登録 {format(new Date(event.created_at), 'yyyy/MM/dd')}
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
