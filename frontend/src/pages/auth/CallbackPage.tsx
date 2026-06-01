import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { userManager } from '../../auth/oidcClient'
import { useAuthStore } from '../../auth/authStore'
import { applyOshiTheme } from '../../lib/applyTheme'
import { Skeleton } from '../../components/ui/Skeleton'

const DEFAULT_OSHI_COLOR = '#FF69B4'
const ME_ENDPOINT = `${import.meta.env.VITE_API_BASE_URL ?? '/api/v1'}/me`

interface MeResponse {
  sub: string
  email: string
  display_name: string
  oshi_color: string
  scopes: string[]
  can_write: boolean
  can_admin: boolean
}

async function fetchMe(accessToken: string): Promise<MeResponse> {
  const res = await fetch(ME_ENDPOINT, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'X-ID-Token': useAuthStore.getState().idToken ?? '',
    },
  })
  if (!res.ok) {
    throw new Error(`/me request failed: ${res.status}`)
  }
  return (await res.json()) as MeResponse
}

export default function CallbackPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)
  const [error, setError] = useState<string | null>(null)
  const hasRunRef = useRef(false)

  useEffect(() => {
    if (hasRunRef.current) return
    hasRunRef.current = true

    async function processCallback() {
      try {
        const user = await userManager.signinRedirectCallback()
        if (!user || !user.access_token) {
          throw new Error('No access token returned from idol-auth.')
        }
        if (!user.id_token) {
          throw new Error('No ID token returned from idol-auth.')
        }

        useAuthStore.setState({
          accessToken: user.access_token,
          idToken: user.id_token,
        })
        const me = await fetchMe(user.access_token)
        const oshiColor = me.oshi_color || DEFAULT_OSHI_COLOR

        setAuth(
          user.access_token,
          user.id_token,
          me.email,
          me.display_name,
          oshiColor,
          me.can_write,
          me.can_admin,
        )
        applyOshiTheme(oshiColor)
        navigate('/idols', { replace: true })
      } catch (err) {
        const message =
          err instanceof Error ? err.message : 'Unknown callback error'
        setError(message)
      }
    }

    void processCallback()
  }, [navigate, setAuth])

  if (error) {
    return (
      <div
        style={{
          minHeight: '100dvh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: 'var(--space-6)',
          background: 'var(--color-bg)',
        }}
      >
        <div
          style={{
            maxWidth: '420px',
            width: '100%',
            padding: 'var(--space-8)',
            background: 'var(--color-surface)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-lg)',
            display: 'flex',
            flexDirection: 'column',
            gap: 'var(--space-4)',
          }}
        >
          <h1
            style={{
              fontSize: 'var(--text-lg)',
              fontWeight: 700,
              color: 'var(--color-text)',
              margin: 0,
            }}
          >
            サインインに失敗しました
          </h1>
          <p
            style={{
              fontSize: 'var(--text-sm)',
              color: 'var(--color-text-muted)',
              margin: 0,
            }}
          >
            {error}
          </p>
          <button
            type="button"
            onClick={() => navigate('/login', { replace: true })}
            style={{
              padding: 'var(--space-3) var(--space-4)',
              background: 'var(--color-accent)',
              color: 'white',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              fontWeight: 600,
              fontSize: 'var(--text-base)',
              cursor: 'pointer',
            }}
          >
            ログイン画面に戻る
          </button>
        </div>
      </div>
    )
  }

  return (
    <div
      style={{
        minHeight: '100dvh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 'var(--space-6)',
        background: 'var(--color-bg)',
      }}
    >
      <div
        style={{
          maxWidth: '400px',
          width: '100%',
          padding: 'var(--space-8)',
          display: 'flex',
          flexDirection: 'column',
          gap: 'var(--space-4)',
        }}
      >
        <Skeleton height="1.5rem" width="60%" />
        <Skeleton height="1rem" />
        <Skeleton height="1rem" width="80%" />
        <p
          style={{
            marginTop: 'var(--space-4)',
            fontSize: 'var(--text-sm)',
            color: 'var(--color-text-muted)',
            textAlign: 'center',
          }}
        >
          サインイン処理中…
        </p>
      </div>
    </div>
  )
}
