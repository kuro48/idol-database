import { useQuery } from '@tanstack/react-query'
import { Skeleton } from '../../components/ui/Skeleton'
import styles from './terms.module.css'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

interface TermsResponse {
  type: string
  content: string
  format: string
}

type TermsType = 'service' | 'privacy'

const PAGE_TITLE: Record<TermsType, string> = {
  service: '利用規約',
  privacy: 'プライバシーポリシー',
}

const API_PATH: Record<TermsType, string> = {
  service: '/terms/service',
  privacy: '/terms/privacy',
}

// Simple markdown-to-JSX renderer for controlled legal document content.
// Handles: # headings, **bold**, numbered lists, bullet lists, paragraphs.
function renderMarkdown(md: string): React.ReactNode[] {
  return md
    .split(/\n{2,}/)
    .map((block, i) => {
      const trimmed = block.trim()
      if (!trimmed) return null

      if (trimmed.startsWith('### ')) return <h3 key={i}>{renderInline(trimmed.slice(4))}</h3>
      if (trimmed.startsWith('## ')) return <h2 key={i}>{renderInline(trimmed.slice(3))}</h2>
      if (trimmed.startsWith('# ')) return <h1 key={i}>{renderInline(trimmed.slice(2))}</h1>

      const lines = trimmed.split('\n').filter(Boolean)

      if (lines.every((l) => /^\d+\.\s/.test(l))) {
        return (
          <ol key={i}>
            {lines.map((l, j) => (
              <li key={j}>{renderInline(l.replace(/^\d+\.\s/, ''))}</li>
            ))}
          </ol>
        )
      }

      if (lines.every((l) => /^[-*]\s/.test(l))) {
        return (
          <ul key={i}>
            {lines.map((l, j) => (
              <li key={j}>{renderInline(l.replace(/^[-*]\s/, ''))}</li>
            ))}
          </ul>
        )
      }

      return <p key={i}>{renderInline(trimmed)}</p>
    })
    .filter((node): node is React.ReactElement => node !== null)
}

function renderInline(text: string): React.ReactNode {
  const parts = text.split(/(\*\*[^*]+\*\*)/)
  if (parts.length === 1) return text
  return parts.map((part, i) =>
    part.startsWith('**') && part.endsWith('**') ? (
      <strong key={i}>{part.slice(2, -2)}</strong>
    ) : (
      part
    ),
  )
}

interface TermsPageProps {
  type: TermsType
}

export default function TermsPage({ type }: TermsPageProps) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['terms', type],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}${API_PATH[type]}`)
      if (!res.ok) throw new Error(`failed: ${res.status}`)
      return (await res.json()) as TermsResponse
    },
    staleTime: 1000 * 60 * 60,
  })

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <h1 className={styles.pageTitle}>{PAGE_TITLE[type]}</h1>

        {isLoading && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <Skeleton height="2rem" width="300px" />
            <Skeleton height="1rem" />
            <Skeleton height="1rem" width="80%" />
            <Skeleton height="1rem" width="60%" />
          </div>
        )}

        {isError && (
          <p style={{ color: 'var(--color-danger)', fontSize: 'var(--text-sm)' }}>
            コンテンツの読み込みに失敗しました。
          </p>
        )}

        {data && <div className={styles.prose}>{renderMarkdown(data.content)}</div>}
      </div>
    </div>
  )
}
