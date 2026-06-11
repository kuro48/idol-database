import { useState } from 'react'
import { useAuth } from '../../auth/useAuth'
import { getValidAuthHeaders } from '../../auth/tokenRefresh'
import styles from './request.module.css'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export default function SubmissionRequestPage() {
  const { isLoggedIn } = useAuth()
  const [targetType, setTargetType] = useState('idol')
  const [payload, setPayload] = useState('{\n  "name": ""\n}')
  const [sourceUrls, setSourceUrls] = useState('')
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!isLoggedIn) {
      setError('申請にはサインインが必要です。')
      return
    }
    setIsSubmitting(true)
    setError(null)
    setMessage(null)
    try {
      const parsedPayload = JSON.parse(payload) as Record<string, unknown>
      const urls = sourceUrls
        .split('\n')
        .map((url) => url.trim())
        .filter(Boolean)
      const { accessToken, idToken } = await getValidAuthHeaders()
      if (!accessToken || !idToken) {
        throw new Error('サインイン情報を更新できませんでした。')
      }
      const res = await fetch(`${API_BASE_URL}/submissions`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'X-ID-Token': idToken,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target_type: targetType,
          payload: parsedPayload,
          source_urls: urls,
        }),
      })
      if (!res.ok) throw new Error(`申請に失敗しました: ${res.status}`)
      const body = (await res.json()) as { submission?: { id?: string } }
      setMessage(`投稿申請を受け付けました。ID: ${body.submission?.id ?? '-'}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '申請に失敗しました。')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>登録申請</h1>
      </div>
      <section className={styles.panel}>
        <form className={styles.form} onSubmit={handleSubmit}>
          <label className={styles.field}>
            <span className={styles.label}>対象</span>
            <select className={styles.select} value={targetType} onChange={(e) => setTargetType(e.target.value)}>
              <option value="idol">アイドル</option>
              <option value="group">グループ</option>
              <option value="agency">事務所</option>
              <option value="event">イベント</option>
            </select>
          </label>
          <label className={styles.field}>
            <span className={styles.label}>登録内容 JSON</span>
            <textarea className={styles.textarea} value={payload} onChange={(e) => setPayload(e.target.value)} />
          </label>
          <label className={styles.field}>
            <span className={styles.label}>参照元 URL</span>
            <textarea
              className={styles.textarea}
              value={sourceUrls}
              onChange={(e) => setSourceUrls(e.target.value)}
              placeholder="https://example.com/source"
            />
          </label>
          {error && <div className={styles.error}>{error}</div>}
          {message && <div className={styles.success}>{message}</div>}
          <button className={styles.button} type="submit" disabled={isSubmitting || !isLoggedIn}>
            {isSubmitting ? '送信中' : '申請する'}
          </button>
        </form>
      </section>
    </div>
  )
}
