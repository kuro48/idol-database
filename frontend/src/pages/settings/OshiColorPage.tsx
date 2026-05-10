import { useState } from 'react'
import { useAuth } from '../../auth/useAuth'
import { applyOshiTheme } from '../../lib/applyTheme'
import { useAuthStore } from '../../auth/authStore'
import styles from '../idols/idol-list.module.css'

const DEFAULT_COLOR = '#FF69B4'

async function updateOshiColor(
  accessToken: string,
  color: string,
): Promise<void> {
  const res = await fetch('/api/v1/me/oshi-color', {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ oshi_color: color }),
  })
  if (!res.ok) throw new Error(`Failed to update color: ${res.status}`)
}

export default function OshiColorPage() {
  const { oshiColor, accessToken, isLoggedIn } = useAuth()
  const setOshiColor = useAuthStore((s) => s.setOshiColor)
  const [color, setColor] = useState(oshiColor ?? DEFAULT_COLOR)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    if (!isLoggedIn || !accessToken) {
      setError('You must be signed in to update your oshi color.')
      return
    }
    setIsSaving(true)
    setError(null)
    setSaved(false)
    try {
      await updateOshiColor(accessToken, color)
      setOshiColor(color)
      applyOshiTheme(color)
      setSaved(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save color.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>推しメンカラー</h1>
      </div>

      <div
        style={{
          background: 'var(--color-surface)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-md)',
          padding: 'var(--space-8)',
          maxWidth: '420px',
          display: 'flex',
          flexDirection: 'column',
          gap: 'var(--space-6)',
        }}
      >
        <p style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text-muted)' }}>
          推しメンカラーはデータ閲覧画面全体のアクセントカラーとして使用されます。
        </p>

        <form onSubmit={handleSave} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            <label htmlFor="oshi-color-picker" style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>
              カラーを選択
            </label>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
              <input
                id="oshi-color-picker"
                type="color"
                value={color}
                onChange={(e) => {
                  setColor(e.target.value)
                  applyOshiTheme(e.target.value)
                }}
                style={{ width: '48px', height: '48px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', cursor: 'pointer', padding: '2px' }}
              />
              <input
                type="text"
                value={color}
                onChange={(e) => {
                  setColor(e.target.value)
                  if (/^#[0-9A-Fa-f]{6}$/.test(e.target.value)) {
                    applyOshiTheme(e.target.value)
                  }
                }}
                placeholder="#FF69B4"
                style={{
                  flex: 1,
                  padding: 'var(--space-2) var(--space-3)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-sm)',
                  fontFamily: 'var(--font-mono)',
                  fontSize: 'var(--text-sm)',
                }}
              />
            </div>
          </div>

          <div
            style={{
              height: '40px',
              borderRadius: 'var(--radius-sm)',
              background: color,
              border: '1px solid var(--color-border)',
            }}
            aria-label="カラープレビュー"
          />

          {error && (
            <div className={styles.error} role="alert">{error}</div>
          )}

          {saved && (
            <div style={{ fontSize: 'var(--text-sm)', color: 'var(--color-success)' }}>
              カラーを保存しました。
            </div>
          )}

          <button
            type="submit"
            disabled={isSaving || !isLoggedIn}
            style={{
              padding: 'var(--space-2) var(--space-4)',
              background: 'var(--color-accent)',
              color: 'white',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              fontWeight: 600,
              fontSize: 'var(--text-base)',
              cursor: isSaving || !isLoggedIn ? 'not-allowed' : 'pointer',
              opacity: isSaving || !isLoggedIn ? 0.5 : 1,
            }}
          >
            {isSaving ? '保存中…' : 'カラーを保存'}
          </button>

          {!isLoggedIn && (
            <p style={{ fontSize: 'var(--text-sm)', color: 'var(--color-text-muted)' }}>
              カラーを保存するにはサインインしてください。
            </p>
          )}
        </form>
      </div>
    </div>
  )
}
