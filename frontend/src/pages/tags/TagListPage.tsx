import { useQuery } from '@tanstack/react-query'
import { Tag } from 'lucide-react'
import { Skeleton } from '../../components/ui/Skeleton'
import styles from '../idols/idol-list.module.css'

interface TagItem {
  id: string
  name: string
  count?: number
}

interface TagsResponse {
  data: TagItem[]
  meta?: { total: number; page: number; per_page: number }
}

async function fetchTags(): Promise<TagsResponse> {
  const res = await fetch('/api/v1/tags?per_page=200')
  if (!res.ok) throw new Error(`Failed to fetch tags: ${res.status}`)
  return res.json() as Promise<TagsResponse>
}

export default function TagListPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['tags'],
    queryFn: fetchTags,
  })

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <div>
          <h1 className={styles.title}>タグ</h1>
          {data?.meta && (
            <p className={styles.count}>{data.meta.total.toLocaleString('ja-JP')} 件</p>
          )}
        </div>
      </div>

      {isError && (
        <div className={styles.error} role="alert">
          タグの読み込みに失敗しました。もう一度お試しください。
        </div>
      )}

      {isLoading ? (
        <div className={styles.tagCloud}>
          {Array.from({ length: 20 }).map((_, i) => (
            <Skeleton key={i} width={`${60 + (i % 5) * 20}px`} height="32px" />
          ))}
        </div>
      ) : (
        <div className={styles.tagCloud}>
          {(data?.data ?? []).map((tag) => (
            <span className={styles.tagPill} key={tag.id}>
              <Tag size={12} aria-hidden="true" />
              {tag.name}
              {tag.count != null && (
                <span className={styles.tagCount}>{tag.count.toLocaleString('ja-JP')}</span>
              )}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
