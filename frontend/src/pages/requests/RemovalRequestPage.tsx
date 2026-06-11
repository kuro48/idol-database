import { useState } from 'react'
import { useAuth } from '../../auth/useAuth'
import { getValidAuthHeaders } from '../../auth/tokenRefresh'
import styles from './request.module.css'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export default function RemovalRequestPage() {
  const { isLoggedIn } = useAuth()
  const [targetType, setTargetType] = useState('idol')
  const [targetId, setTargetId] = useState('')
  const [requesterType, setRequesterType] = useState('third_party')
  const [reason, setReason] = useState('')
  const [evidence, setEvidence] = useState('')
  const [description, setDescription] = useState('')
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
      const { accessToken, idToken } = await getValidAuthHeaders()
      if (!accessToken || !idToken) {
        throw new Error('サインイン情報を更新できませんでした。')
      }
      const res = await fetch(`${API_BASE_URL}/removal-requests`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'X-ID-Token': idToken,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target_type: targetType,
          target_id: targetId,
          requester_type: requesterType,
          reason,
          evidence: evidence || undefined,
          description,
        }),
      })
      if (!res.ok) throw new Error(`申請に失敗しました: ${res.status}`)
      const body = (await res.json()) as { removal_request?: { id?: string } }
      setMessage(`削除申請を受け付けました。ID: ${body.removal_request?.id ?? '-'}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '申請に失敗しました。')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <h1 className={styles.title}>削除申請</h1>
      </div>
      <section className={styles.panel}>
        <form className={styles.form} onSubmit={handleSubmit}>
          <label className={styles.field}>
            <span className={styles.label}>対象</span>
            <select className={styles.select} value={targetType} onChange={(e) => setTargetType(e.target.value)}>
              <option value="idol">アイドル</option>
              <option value="group">グループ</option>
            </select>
          </label>
          <label className={styles.field}>
            <span className={styles.label}>対象 ID</span>
            <input className={styles.input} value={targetId} onChange={(e) => setTargetId(e.target.value)} required />
          </label>
          <label className={styles.field}>
            <span className={styles.label}>申請者種別</span>
            <select className={styles.select} value={requesterType} onChange={(e) => setRequesterType(e.target.value)}>
              <option value="third_party">第三者</option>
              <option value="idol_themself">本人</option>
              <option value="agency">事務所</option>
            </select>
          </label>
          <label className={styles.field}>
            <span className={styles.label}>理由</span>
            <textarea className={styles.textarea} value={reason} onChange={(e) => setReason(e.target.value)} required />
          </label>
          <label className={styles.field}>
            <span className={styles.label}>証拠 URL</span>
            <input className={styles.input} value={evidence} onChange={(e) => setEvidence(e.target.value)} />
          </label>
          <label className={styles.field}>
            <span className={styles.label}>詳細</span>
            <textarea className={styles.textarea} value={description} onChange={(e) => setDescription(e.target.value)} required />
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
