import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { format } from 'date-fns'
import { CheckCircle2, XCircle, AlertTriangle, Clock } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { useAuthStore } from '../../auth/authStore'
import { getValidAuthHeaders } from '../../auth/tokenRefresh'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import { DataTable } from '../../components/table/DataTable'
import styles from './dashboard.module.css'
import reviewStyles from './review.module.css'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

interface Submission {
  id: string
  target_type: string
  payload: string
  source_urls: string[]
  contributor_email: string
  status: string
  revision_note: string
  created_at: string
}

interface RemovalRequest {
  id: string
  target_id: string
  target_type: string
  requester_type: string
  reason: string
  contact_info: string
  evidence: string
  description: string
  status: string
  created_at: string
  sla_due_at: string
  sla_overdue: boolean
}

type SubmissionStatus = 'approved' | 'rejected' | 'needs_revision'
type RemovalStatus = 'approved' | 'rejected'
type Tab = 'submissions' | 'removals'

const TARGET_LABEL: Record<string, string> = {
  idol: 'アイドル',
  group: 'グループ',
  agency: '事務所',
  event: 'イベント',
}

const REQUESTER_LABEL: Record<string, string> = {
  idol_themself: '本人',
  agency: '事務所',
  third_party: '第三者',
}

type BadgeVariant = 'warning' | 'success' | 'danger' | 'info' | 'default'

function statusVariant(status: string): BadgeVariant {
  if (status === 'pending') return 'warning'
  if (status === 'approved') return 'success'
  if (status === 'rejected') return 'danger'
  if (status === 'needs_revision') return 'info'
  return 'default'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '審査待ち',
    approved: '承認',
    rejected: '却下',
    needs_revision: '要修正',
  }
  return map[status] ?? status
}

async function adminAuthHeaders() {
  const { accessToken } = await getValidAuthHeaders()
  if (!accessToken) throw new Error('authentication required')
  return { Authorization: `Bearer ${accessToken}` }
}

async function adminGet<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: await adminAuthHeaders(),
  })
  if (!res.ok) throw new Error(`request failed: ${res.status}`)
  return (await res.json()) as T
}

async function adminPut(path: string, body: unknown): Promise<void> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'PUT',
    headers: { ...(await adminAuthHeaders()), 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = (await res.json().catch(() => null)) as { message?: string } | null
    throw new Error(err?.message ?? `request failed: ${res.status}`)
  }
}

export default function ReviewPage() {
  const email = useAuthStore((s) => s.email) ?? ''
  const queryClient = useQueryClient()

  const [tab, setTab] = useState<Tab>('submissions')
  const [showAll, setShowAll] = useState(false)

  const [selectedSub, setSelectedSub] = useState<Submission | null>(null)
  const [subStatus, setSubStatus] = useState<SubmissionStatus>('approved')
  const [subNote, setSubNote] = useState('')
  const [subError, setSubError] = useState<string | null>(null)

  const [selectedRemoval, setSelectedRemoval] = useState<RemovalRequest | null>(null)
  const [removalStatus, setRemovalStatus] = useState<RemovalStatus>('approved')
  const [removalError, setRemovalError] = useState<string | null>(null)

  const { data: subData, isLoading: subLoading } = useQuery({
    queryKey: ['admin', 'submissions', showAll],
    queryFn: () =>
      adminGet<{ submissions: Submission[]; count: number }>(
        showAll ? '/submissions' : '/submissions/pending',
      ),
    enabled: tab === 'submissions',
  })

  const { data: removalData, isLoading: removalLoading } = useQuery({
    queryKey: ['admin', 'removals', showAll],
    queryFn: () =>
      adminGet<{ removal_requests: RemovalRequest[]; count: number }>(
        showAll ? '/removal-requests' : '/removal-requests/pending',
      ),
    enabled: tab === 'removals',
  })

  const updateSubMutation = useMutation({
    mutationFn: ({ id, status, note }: { id: string; status: SubmissionStatus; note: string }) =>
      adminPut(
        `/submissions/${id}/status`,
        { status, reviewed_by: email, revision_note: note },
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['admin', 'submissions'] })
      setSelectedSub(null)
      setSubNote('')
      setSubError(null)
    },
    onError: (err: Error) => setSubError(err.message),
  })

  const updateRemovalMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: RemovalStatus }) =>
      adminPut(`/removal-requests/${id}`, { status }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['admin', 'removals'] })
      setSelectedRemoval(null)
      setRemovalError(null)
    },
    onError: (err: Error) => setRemovalError(err.message),
  })

  const subColumns: ColumnDef<Submission, unknown>[] = [
    {
      accessorKey: 'target_type',
      header: '種別',
      cell: ({ getValue }) => TARGET_LABEL[getValue<string>()] ?? getValue<string>(),
    },
    {
      accessorKey: 'contributor_email',
      header: '投稿者',
      cell: ({ getValue }) => (
        <span style={{ color: 'var(--color-text-muted)', fontSize: 'var(--text-sm)' }}>
          {getValue<string>() || '—'}
        </span>
      ),
    },
    {
      accessorKey: 'status',
      header: 'ステータス',
      cell: ({ getValue }) => (
        <Badge variant={statusVariant(getValue<string>())}>{statusLabel(getValue<string>())}</Badge>
      ),
    },
    {
      accessorKey: 'created_at',
      header: '申請日',
      cell: ({ getValue }) => format(new Date(getValue<string>()), 'yyyy/MM/dd HH:mm'),
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            setSelectedSub(row.original)
            setSubStatus('approved')
            setSubNote('')
            setSubError(null)
          }}
        >
          審査
        </Button>
      ),
    },
  ]

  const removalColumns: ColumnDef<RemovalRequest, unknown>[] = [
    {
      accessorKey: 'target_type',
      header: '種別',
      cell: ({ getValue }) => TARGET_LABEL[getValue<string>()] ?? getValue<string>(),
    },
    {
      accessorKey: 'target_id',
      header: '対象ID',
      cell: ({ getValue }) => (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--color-text-muted)' }}>
          {getValue<string>()}
        </span>
      ),
    },
    {
      accessorKey: 'requester_type',
      header: '申請者',
      cell: ({ getValue }) => REQUESTER_LABEL[getValue<string>()] ?? getValue<string>(),
    },
    {
      accessorKey: 'status',
      header: 'ステータス',
      cell: ({ getValue }) => (
        <Badge variant={statusVariant(getValue<string>())}>{statusLabel(getValue<string>())}</Badge>
      ),
    },
    {
      accessorKey: 'sla_overdue',
      header: 'SLA',
      cell: ({ getValue }) =>
        getValue<boolean>() ? (
          <span
            style={{
              color: 'var(--color-danger)',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              fontSize: 'var(--text-xs)',
            }}
          >
            <AlertTriangle size={12} aria-hidden="true" />
            超過
          </span>
        ) : null,
    },
    {
      accessorKey: 'created_at',
      header: '申請日',
      cell: ({ getValue }) => format(new Date(getValue<string>()), 'yyyy/MM/dd HH:mm'),
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            setSelectedRemoval(row.original)
            setRemovalStatus('approved')
            setRemovalError(null)
          }}
        >
          審査
        </Button>
      ),
    },
  ]

  return (
    <div className={styles.page}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: 'var(--space-4)',
        }}
      >
        <h1 className={styles.heading}>審査</h1>
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 'var(--space-2)',
            fontSize: 'var(--text-sm)',
            color: 'var(--color-text-muted)',
            cursor: 'pointer',
          }}
        >
          <input type="checkbox" checked={showAll} onChange={(e) => setShowAll(e.target.checked)} />
          すべて表示
        </label>
      </div>

      <div className={reviewStyles.tabs} role="tablist">
        {(
          [
            { key: 'submissions', label: '投稿申請', count: subData?.count, icon: Clock },
            { key: 'removals', label: '削除申請', count: removalData?.count, icon: XCircle },
          ] as const
        ).map(({ key, label, count, icon: Icon }) => (
          <button
            key={key}
            role="tab"
            aria-selected={tab === key}
            className={`${reviewStyles.tab} ${tab === key ? reviewStyles.tabActive : ''}`}
            onClick={() => setTab(key)}
            type="button"
          >
            <Icon size={14} aria-hidden="true" />
            {label}
            {count != null && count > 0 && (
              <span className={reviewStyles.badge}>{count}</span>
            )}
          </button>
        ))}
      </div>

      {tab === 'submissions' && (
        <DataTable
          columns={subColumns}
          data={subData?.submissions ?? []}
          isLoading={subLoading}
          emptyMessage="審査待ちの投稿申請はありません"
        />
      )}

      {tab === 'removals' && (
        <DataTable
          columns={removalColumns}
          data={removalData?.removal_requests ?? []}
          isLoading={removalLoading}
          emptyMessage="審査待ちの削除申請はありません"
        />
      )}

      {/* Submission review modal */}
      <Modal isOpen={selectedSub !== null} onClose={() => setSelectedSub(null)} title="投稿申請の審査">
        {selectedSub && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <dl className={reviewStyles.detailGrid}>
              <dt>申請ID</dt>
              <dd style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{selectedSub.id}</dd>
              <dt>種別</dt>
              <dd>{TARGET_LABEL[selectedSub.target_type] ?? selectedSub.target_type}</dd>
              <dt>投稿者</dt>
              <dd>{selectedSub.contributor_email || '—'}</dd>
              <dt>申請日</dt>
              <dd>{format(new Date(selectedSub.created_at), 'yyyy/MM/dd HH:mm')}</dd>
            </dl>

            <div>
              <p className={reviewStyles.sectionLabel}>内容 (JSON)</p>
              <pre className={reviewStyles.payload}>{selectedSub.payload}</pre>
            </div>

            {selectedSub.source_urls.length > 0 && (
              <div>
                <p className={reviewStyles.sectionLabel}>参照元URL</p>
                <ul style={{ paddingLeft: 'var(--space-4)', display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                  {selectedSub.source_urls.map((url) => (
                    <li key={url} style={{ fontSize: 'var(--text-sm)' }}>
                      <a href={url} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--color-accent)' }}>
                        {url}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              <p style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>判定</p>
              <div style={{ display: 'flex', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
                {(
                  [
                    { value: 'approved', icon: CheckCircle2 },
                    { value: 'rejected', icon: XCircle },
                    { value: 'needs_revision', icon: AlertTriangle },
                  ] as const
                ).map(({ value, icon: Icon }) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setSubStatus(value)}
                    className={reviewStyles.statusChip}
                    data-active={subStatus === value}
                    data-variant={statusVariant(value)}
                  >
                    <Icon size={13} aria-hidden="true" />
                    {statusLabel(value)}
                  </button>
                ))}
              </div>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              <label htmlFor="sub-note" style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>
                コメント
                {subStatus === 'needs_revision' && (
                  <span style={{ color: 'var(--color-danger)', marginLeft: 4 }}>*</span>
                )}
              </label>
              <textarea
                id="sub-note"
                value={subNote}
                onChange={(e) => setSubNote(e.target.value)}
                placeholder="差し戻し理由や補足など…"
                rows={3}
                style={{
                  padding: 'var(--space-2) var(--space-3)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-sm)',
                  fontSize: 'var(--text-sm)',
                  resize: 'vertical',
                }}
              />
            </div>

            {subError && (
              <div role="alert" style={{ color: 'var(--color-danger)', fontSize: 'var(--text-sm)' }}>
                {subError}
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
              <Button variant="secondary" size="sm" onClick={() => setSelectedSub(null)}>
                キャンセル
              </Button>
              <Button
                size="sm"
                variant={subStatus === 'rejected' ? 'danger' : 'primary'}
                disabled={updateSubMutation.isPending}
                onClick={() => {
                  if (subStatus === 'needs_revision' && !subNote.trim()) {
                    setSubError('要修正の場合はコメントが必要です')
                    return
                  }
                  updateSubMutation.mutate({ id: selectedSub.id, status: subStatus, note: subNote })
                }}
              >
                {updateSubMutation.isPending ? '送信中…' : '確定'}
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Removal review modal */}
      <Modal
        isOpen={selectedRemoval !== null}
        onClose={() => setSelectedRemoval(null)}
        title="削除申請の審査"
      >
        {selectedRemoval && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <dl className={reviewStyles.detailGrid}>
              <dt>申請ID</dt>
              <dd style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{selectedRemoval.id}</dd>
              <dt>対象種別</dt>
              <dd>{TARGET_LABEL[selectedRemoval.target_type] ?? selectedRemoval.target_type}</dd>
              <dt>対象ID</dt>
              <dd style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{selectedRemoval.target_id}</dd>
              <dt>申請者種別</dt>
              <dd>{REQUESTER_LABEL[selectedRemoval.requester_type] ?? selectedRemoval.requester_type}</dd>
              <dt>連絡先</dt>
              <dd>{selectedRemoval.contact_info || '—'}</dd>
              <dt>申請日</dt>
              <dd>{format(new Date(selectedRemoval.created_at), 'yyyy/MM/dd HH:mm')}</dd>
              <dt>SLA期限</dt>
              <dd style={{ color: selectedRemoval.sla_overdue ? 'var(--color-danger)' : undefined }}>
                {format(new Date(selectedRemoval.sla_due_at), 'yyyy/MM/dd HH:mm')}
                {selectedRemoval.sla_overdue && ' (超過)'}
              </dd>
            </dl>

            <div>
              <p className={reviewStyles.sectionLabel}>理由</p>
              <p
                style={{
                  fontSize: 'var(--text-sm)',
                  padding: 'var(--space-3)',
                  background: 'var(--color-surface-2)',
                  borderRadius: 'var(--radius-sm)',
                }}
              >
                {selectedRemoval.reason}
              </p>
            </div>

            <div>
              <p className={reviewStyles.sectionLabel}>詳細</p>
              <p
                style={{
                  fontSize: 'var(--text-sm)',
                  padding: 'var(--space-3)',
                  background: 'var(--color-surface-2)',
                  borderRadius: 'var(--radius-sm)',
                }}
              >
                {selectedRemoval.description}
              </p>
            </div>

            {selectedRemoval.evidence && (
              <div>
                <p className={reviewStyles.sectionLabel}>根拠URL</p>
                <a
                  href={selectedRemoval.evidence}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{ fontSize: 'var(--text-sm)', color: 'var(--color-accent)', wordBreak: 'break-all' }}
                >
                  {selectedRemoval.evidence}
                </a>
              </div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              <p style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>判定</p>
              <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
                {(
                  [
                    { value: 'approved', icon: CheckCircle2 },
                    { value: 'rejected', icon: XCircle },
                  ] as const
                ).map(({ value, icon: Icon }) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setRemovalStatus(value)}
                    className={reviewStyles.statusChip}
                    data-active={removalStatus === value}
                    data-variant={statusVariant(value)}
                  >
                    <Icon size={13} aria-hidden="true" />
                    {statusLabel(value)}
                  </button>
                ))}
              </div>
            </div>

            {removalError && (
              <div role="alert" style={{ color: 'var(--color-danger)', fontSize: 'var(--text-sm)' }}>
                {removalError}
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
              <Button variant="secondary" size="sm" onClick={() => setSelectedRemoval(null)}>
                キャンセル
              </Button>
              <Button
                size="sm"
                variant={removalStatus === 'rejected' ? 'danger' : 'primary'}
                disabled={updateRemovalMutation.isPending}
                onClick={() =>
                  updateRemovalMutation.mutate({ id: selectedRemoval.id, status: removalStatus })
                }
              >
                {updateRemovalMutation.isPending ? '送信中…' : '確定'}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
