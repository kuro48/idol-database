import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '../ui/Button'

interface PaginationProps {
  page: number
  perPage: number
  total: number
  onPageChange: (page: number) => void
  onPerPageChange?: (perPage: number) => void
}

const PER_PAGE_OPTIONS = [10, 20, 50, 100]

export function Pagination({
  page,
  perPage,
  total,
  onPageChange,
  onPerPageChange,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / perPage))
  const start = total === 0 ? 0 : (page - 1) * perPage + 1
  const end = Math.min(page * perPage, total)

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: 'var(--space-3) var(--space-4)',
        borderTop: '1px solid var(--color-border)',
        fontSize: 'var(--text-sm)',
        color: 'var(--color-text-muted)',
        flexWrap: 'wrap',
        gap: 'var(--space-3)',
      }}
    >
      <span>
        {total === 0
          ? 'No results'
          : `${start}–${end} of ${total.toLocaleString()}`}
      </span>

      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        {onPerPageChange && (
          <label
            style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}
          >
            Per page:
            <select
              value={perPage}
              onChange={(e) => onPerPageChange(Number(e.target.value))}
              style={{
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-sm)',
                padding: '2px var(--space-2)',
                fontSize: 'var(--text-sm)',
                background: 'var(--color-surface)',
                color: 'var(--color-text)',
              }}
            >
              {PER_PAGE_OPTIONS.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
        )}

        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            aria-label="Previous page"
          >
            <ChevronLeft size={14} />
          </Button>

          <span style={{ padding: '0 var(--space-2)' }}>
            {page} / {totalPages}
          </span>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            aria-label="Next page"
          >
            <ChevronRight size={14} />
          </Button>
        </div>
      </div>
    </div>
  )
}
