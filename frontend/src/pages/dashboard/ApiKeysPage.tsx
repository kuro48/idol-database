import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { format } from 'date-fns'
import { Search, Plus, Trash2, Copy, Check } from 'lucide-react'
import { getValidAuthHeaders } from '../../auth/tokenRefresh'
import { Button } from '../../components/ui/Button'
import { Modal } from '../../components/ui/Modal'
import styles from './dashboard.module.css'
import listStyles from '../idols/idol-list.module.css'

interface APIKey {
  id: string
  masked_key: string
  email: string
  name: string
  plan_type: string
  is_active: boolean
  created_at: string
}

interface CreateKeyResponse extends APIKey {
  raw_key: string
}

async function adminHeaders() {
  const { accessToken } = await getValidAuthHeaders()
  if (!accessToken) throw new Error('authentication required')
  return { Authorization: `Bearer ${accessToken}` }
}

async function listKeys(email: string): Promise<APIKey[]> {
  const res = await fetch(`/api/v1/admin/apikeys?email=${encodeURIComponent(email)}`, {
    headers: await adminHeaders(),
  })
  if (!res.ok) throw new Error(`Failed: ${res.status}`)
  return res.json() as Promise<APIKey[]>
}

async function createKey(
  data: { email: string; name: string; plan_type: string },
): Promise<CreateKeyResponse> {
  const res = await fetch('/api/v1/admin/apikeys', {
    method: 'POST',
    headers: { ...(await adminHeaders()), 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) {
    const err = (await res.json().catch(() => null)) as { message?: string } | null
    throw new Error(err?.message ?? `Failed: ${res.status}`)
  }
  return res.json() as Promise<CreateKeyResponse>
}

async function revokeKey(id: string): Promise<void> {
  const res = await fetch(`/api/v1/admin/apikeys/${id}`, {
    method: 'DELETE',
    headers: await adminHeaders(),
  })
  if (!res.ok && res.status !== 204) throw new Error(`Failed: ${res.status}`)
}

const PLAN_TYPES = ['free', 'developer', 'business'] as const

export default function ApiKeysPage() {
  const queryClient = useQueryClient()

  const [emailInput, setEmailInput] = useState('')
  const [searchEmail, setSearchEmail] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newEmail, setNewEmail] = useState('')
  const [newName, setNewName] = useState('')
  const [newPlan, setNewPlan] = useState<(typeof PLAN_TYPES)[number]>('free')
  const [createError, setCreateError] = useState<string | null>(null)
  const [createdKey, setCreatedKey] = useState<CreateKeyResponse | null>(null)
  const [copied, setCopied] = useState(false)

  const { data: keys, isLoading, isError } = useQuery({
    queryKey: ['admin', 'apikeys', searchEmail],
    queryFn: () => listKeys(searchEmail),
    enabled: searchEmail !== '',
  })

  const createMutation = useMutation({
    mutationFn: (d: { email: string; name: string; plan_type: string }) =>
      createKey(d),
    onSuccess: (data) => {
      setCreatedKey(data)
      setShowCreate(false)
      setNewEmail('')
      setNewName('')
      setNewPlan('free')
      setCreateError(null)
      void queryClient.invalidateQueries({ queryKey: ['admin', 'apikeys', data.email] })
    },
    onError: (err: Error) => {
      setCreateError(err.message)
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => revokeKey(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['admin', 'apikeys', searchEmail] })
    },
  })

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setSearchEmail(emailInput.trim())
  }

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreateError(null)
    createMutation.mutate({ email: newEmail, name: newName, plan_type: newPlan })
  }

  async function handleCopy(text: string) {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className={styles.page}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 'var(--space-4)' }}>
        <h1 className={styles.heading}>APIキー</h1>
        <Button size="sm" onClick={() => { setShowCreate(true); setCreateError(null) }}>
          <Plus size={14} aria-hidden="true" style={{ marginRight: 'var(--space-2)' }} />
          新規作成
        </Button>
      </div>

      <form onSubmit={handleSearch} className={listStyles.searchForm} style={{ maxWidth: '480px' }}>
        <div className={listStyles.searchWrapper}>
          <Search size={14} className={listStyles.searchIcon} aria-hidden="true" />
          <input
            type="email"
            value={emailInput}
            onChange={(e) => setEmailInput(e.target.value)}
            placeholder="メールアドレスで検索…"
            className={listStyles.searchInput}
          />
        </div>
      </form>

      {searchEmail && isError && (
        <div className={listStyles.error} role="alert">APIキーの読み込みに失敗しました。</div>
      )}

      {searchEmail && !isError && (
        <div className={listStyles.tableCard}>
          {isLoading ? (
            <p style={{ padding: 'var(--space-6)', color: 'var(--color-text-muted)', fontSize: 'var(--text-sm)' }}>
              読み込み中…
            </p>
          ) : !keys || keys.length === 0 ? (
            <p style={{ padding: 'var(--space-6)', color: 'var(--color-text-muted)', fontSize: 'var(--text-sm)' }}>
              {searchEmail} のAPIキーは見つかりませんでした。
            </p>
          ) : (
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--text-sm)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                  {['名前', 'マスクキー', 'プラン', 'ステータス', '作成日', ''].map((h) => (
                    <th
                      key={h}
                      style={{ padding: 'var(--space-3) var(--space-4)', textAlign: 'left', color: 'var(--color-text-muted)', fontWeight: 500 }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {keys.map((k) => (
                  <tr key={k.id} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: 'var(--space-3) var(--space-4)' }}>{k.name}</td>
                    <td style={{ padding: 'var(--space-3) var(--space-4)', fontFamily: 'var(--font-mono)', color: 'var(--color-text-muted)' }}>
                      {k.masked_key}
                    </td>
                    <td style={{ padding: 'var(--space-3) var(--space-4)', textTransform: 'capitalize' }}>{k.plan_type}</td>
                    <td style={{ padding: 'var(--space-3) var(--space-4)' }}>
                      <span style={{
                        display: 'inline-block',
                        padding: '2px 8px',
                        borderRadius: 'var(--radius-sm)',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        background: k.is_active
                          ? 'color-mix(in srgb, var(--color-success) 12%, white)'
                          : 'var(--color-surface-2)',
                        color: k.is_active ? 'var(--color-success)' : 'var(--color-text-muted)',
                      }}>
                        {k.is_active ? '有効' : '無効'}
                      </span>
                    </td>
                    <td style={{ padding: 'var(--space-3) var(--space-4)', color: 'var(--color-text-muted)' }}>
                      {format(new Date(k.created_at), 'MMM d, yyyy')}
                    </td>
                    <td style={{ padding: 'var(--space-3) var(--space-4)' }}>
                      {k.is_active && (
                        <button
                          onClick={() => revokeMutation.mutate(k.id)}
                          disabled={revokeMutation.isPending}
                          title="無効化"
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 'var(--space-1)',
                            padding: 'var(--space-1) var(--space-2)',
                            borderRadius: 'var(--radius-sm)',
                            border: '1px solid var(--color-border)',
                            background: 'none',
                            color: 'var(--color-danger)',
                            cursor: 'pointer',
                            fontSize: 'var(--text-sm)',
                          }}
                          type="button"
                        >
                          <Trash2 size={13} aria-hidden="true" />
                          無効化
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="APIキーの新規作成">
        <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          {([
            { id: 'key-email', label: 'メールアドレス', type: 'email', value: newEmail, set: setNewEmail, placeholder: 'user@example.com' },
            { id: 'key-name', label: 'キー名', type: 'text', value: newName, set: setNewName, placeholder: '自分のインテグレーション' },
          ] as const).map(({ id, label, type, value, set, placeholder }) => (
            <div key={id} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              <label htmlFor={id} style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>{label}</label>
              <input
                id={id}
                type={type}
                value={value}
                onChange={(e) => set(e.target.value)}
                placeholder={placeholder}
                required
                style={{ padding: 'var(--space-2) var(--space-3)', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', fontSize: 'var(--text-sm)' }}
              />
            </div>
          ))}

          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
            <label htmlFor="key-plan" style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>プラン</label>
            <select
              id="key-plan"
              value={newPlan}
              onChange={(e) => setNewPlan(e.target.value as (typeof PLAN_TYPES)[number])}
              style={{ padding: 'var(--space-2) var(--space-3)', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', fontSize: 'var(--text-sm)', background: 'var(--color-surface)' }}
            >
              {PLAN_TYPES.map((p) => (
                <option key={p} value={p}>{p}</option>
              ))}
            </select>
          </div>

          {createError && (
            <div className={listStyles.error} role="alert">{createError}</div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)', marginTop: 'var(--space-2)' }}>
            <Button type="button" variant="secondary" size="sm" onClick={() => setShowCreate(false)}>
              キャンセル
            </Button>
            <Button type="submit" size="sm" disabled={createMutation.isPending}>
              {createMutation.isPending ? '作成中…' : '作成'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal isOpen={createdKey !== null} onClose={() => setCreatedKey(null)} title="APIキーを作成しました">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          <p style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text-muted)' }}>
            このキーを今すぐコピーしてください。二度と表示されません。
          </p>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: 'var(--space-3)',
            background: 'var(--color-surface-2)',
            padding: 'var(--space-3) var(--space-4)',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-border)',
          }}>
            <code style={{ flex: 1, fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', wordBreak: 'break-all' }}>
              {createdKey?.raw_key}
            </code>
            <button
              onClick={() => createdKey && void handleCopy(createdKey.raw_key)}
              style={{ flexShrink: 0, padding: 'var(--space-2)', background: 'none', border: 'none', cursor: 'pointer', color: copied ? 'var(--color-success)' : 'var(--color-text-muted)' }}
              type="button"
              aria-label="キーをコピー"
            >
              {copied ? <Check size={16} /> : <Copy size={16} />}
            </button>
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button size="sm" onClick={() => setCreatedKey(null)}>完了</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
