import { useState } from 'react'
import { LogIn, UserPlus } from 'lucide-react'
import { useLocation } from 'react-router-dom'
import { userManager } from './oidcClient'
import styles from './login.module.css'

export default function LoginPage() {
  const location = useLocation()
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const rawReturnTo = new URLSearchParams(location.search).get('return_to')
  const returnTo =
    rawReturnTo && rawReturnTo.startsWith('/') && !rawReturnTo.startsWith('//')
      ? rawReturnTo
      : '/idols'
  const registrationReturnTo = `${window.location.origin}/login?return_to=${encodeURIComponent(returnTo)}`

  async function handleSignIn() {
    setIsLoading(true)
    setError(null)
    try {
      await userManager.signinRedirect({ returnTo })
    } catch (err) {
      setError(
        err instanceof Error
          ? `サインインに失敗しました: ${err.message}`
          : 'サインインに失敗しました。もう一度お試しください。',
      )
      setIsLoading(false)
    }
  }

  async function handleRegistration() {
    setIsLoading(true)
    setError(null)
    try {
      await userManager.registrationRedirect(registrationReturnTo)
    } catch (err) {
      setError(
        err instanceof Error
          ? `アカウント作成に進めませんでした: ${err.message}`
          : 'アカウント作成に進めませんでした。もう一度お試しください。',
      )
      setIsLoading(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles.logo}>idol.db</div>
          <p className={styles.subtitle}>
            idol-auth でサインインして続行してください
          </p>
        </div>

        <div className={styles.form}>
          {error && (
            <div className={styles.error} role="alert">
              {error}
            </div>
          )}

          <button
            type="button"
            onClick={handleSignIn}
            disabled={isLoading}
            className={styles.button}
          >
            <LogIn
              size={16}
              aria-hidden="true"
              style={{ marginRight: '0.5rem', verticalAlign: '-3px' }}
            />
            {isLoading ? 'リダイレクト中…' : 'idol-auth でサインイン'}
          </button>

          <button
            type="button"
            onClick={handleRegistration}
            disabled={isLoading}
            className={styles.secondaryButton}
          >
            <UserPlus
              size={16}
              aria-hidden="true"
              style={{ marginRight: '0.5rem', verticalAlign: '-3px' }}
            />
            アカウントを作成
          </button>
        </div>

        <p className={styles.hint}>
          サインインしなくても公開データは閲覧できます。{' '}
          <a href="/" className={styles.hintLink}>
            ゲストとして続行 →
          </a>
        </p>
      </div>
    </div>
  )
}
