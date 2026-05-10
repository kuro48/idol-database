import { useState } from 'react'
import { LogIn } from 'lucide-react'
import { userManager } from './oidcClient'
import styles from './login.module.css'

export default function LoginPage() {
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  async function handleSignIn() {
    setIsLoading(true)
    setError(null)
    try {
      await userManager.signinRedirect()
    } catch (err) {
      setError(
        err instanceof Error
          ? `サインインに失敗しました: ${err.message}`
          : 'サインインに失敗しました。もう一度お試しください。',
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
