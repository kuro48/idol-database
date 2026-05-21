import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { Search, Disc3, CalendarDays } from 'lucide-react'
import { Skeleton } from '../../components/ui/Skeleton'
import { Pagination } from '../../components/table/Pagination'
import styles from '../idols/idol-list.module.css'

interface Release {
  id: string
  title: string
  type?: string
  release_date?: string
}

interface ReleasesResponse {
  data: Release[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchReleases(page: number, perPage: number, q: string): Promise<ReleasesResponse> {
  const params = new URLSearchParams({
    page: String(page),
    per_page: String(perPage),
    ...(q ? { q } : {}),
  })
  const res = await fetch(`/api/v1/releases?${params}`)
  if (!res.ok) throw new Error(`Failed to fetch releases: ${res.status}`)
  return res.json() as Promise<ReleasesResponse>
}

const SKELETON_COUNT = 12

export default function ReleaseListPage() {
  const [page, setPage] = useState(1)
  const [perPage] = useState(24)
  const [search, setSearch] = useState('')
  const [q, setQ] = useState('')

  const { data, isLoading, isError } = useQuery({
    queryKey: ['releases', page, perPage, q],
    queryFn: () => fetchReleases(page, perPage, q),
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
          <h1 className={styles.title}>リリース</h1>
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
              placeholder="リリースを検索…"
              className={styles.searchInput}
            />
          </div>
        </form>
      </div>

      {isError && (
        <div className={styles.error} role="alert">
          リリースの読み込みに失敗しました。もう一度お試しください。
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
          : (data?.data ?? []).map((release) => (
              <article className={styles.card} key={release.id}>
                <div className={styles.cardIcon} aria-hidden="true">
                  <Disc3 size={20} />
                </div>
                <p className={styles.cardName}>{release.title}</p>
                <div className={styles.cardMeta}>
                  {release.type && (
                    <span className={styles.badge}>{release.type}</span>
                  )}
                  {release.release_date && (
                    <span className={styles.cardMetaItem}>
                      <CalendarDays size={12} aria-hidden="true" />
                      {format(new Date(release.release_date), 'yyyy/MM/dd')}
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
