interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  refresh_token?: string
  id_token?: string
  scope?: string
}

interface LoginState {
  state: string
  codeVerifier: string
}

const AUTH_BASE_URL = (
  import.meta.env.VITE_IDOL_AUTH_BASE_URL ??
  import.meta.env.VITE_OIDC_ISSUER ??
  ''
).replace(/\/$/, '')
const CLIENT_ID = import.meta.env.VITE_OIDC_CLIENT_ID
const REDIRECT_URI = import.meta.env.VITE_OIDC_REDIRECT_URI
const POST_LOGOUT_URI = import.meta.env.VITE_OIDC_POST_LOGOUT_URI
const SCOPE = 'openid email profile'
const STORAGE_KEY = 'idol-db-oidc-login'

function randomBase64Url(bytes: number) {
  const values = new Uint8Array(bytes)
  crypto.getRandomValues(values)
  return base64Url(values)
}

function base64Url(bytes: Uint8Array) {
  let binary = ''
  bytes.forEach((value) => {
    binary += String.fromCharCode(value)
  })
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

async function sha256Base64Url(value: string) {
  const data = new TextEncoder().encode(value)
  const digest = await crypto.subtle.digest('SHA-256', data)
  return base64Url(new Uint8Array(digest))
}

function requireConfig() {
  if (!AUTH_BASE_URL || !CLIENT_ID || !REDIRECT_URI) {
    throw new Error('idol-auth のフロントエンド設定が不足しています。')
  }
}

export const userManager = {
  async signinRedirect() {
    requireConfig()
    const state = randomBase64Url(24)
    const codeVerifier = randomBase64Url(48)
    const codeChallenge = await sha256Base64Url(codeVerifier)
    const loginState: LoginState = { state, codeVerifier }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(loginState))

    const params = new URLSearchParams({
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
      response_type: 'code',
      scope: SCOPE,
      state,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    })
    window.location.assign(`${AUTH_BASE_URL}/v1/public/browser/login?${params}`)
  },

  async signinRedirectCallback(): Promise<TokenResponse> {
    requireConfig()
    const params = new URLSearchParams(window.location.search)
    const error = params.get('error')
    if (error) {
      throw new Error(params.get('error_description') ?? error)
    }
    const code = params.get('code')
    const state = params.get('state')
    const saved = localStorage.getItem(STORAGE_KEY)
    if (!code || !state || !saved) {
      throw new Error('ログインコールバックが不正です。')
    }
    const loginState = JSON.parse(saved) as LoginState
    localStorage.removeItem(STORAGE_KEY)
    if (loginState.state !== state) {
      throw new Error('ログイン state が一致しません。')
    }

    const body = new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: REDIRECT_URI,
      client_id: CLIENT_ID,
      code_verifier: loginState.codeVerifier,
    })
    const res = await fetch(`${AUTH_BASE_URL}/v1/public/api/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body,
    })
    if (!res.ok) {
      throw new Error(`token request failed: ${res.status}`)
    }
    return (await res.json()) as TokenResponse
  },

  async signoutRedirect(idToken?: string | null) {
    const params = new URLSearchParams()
    if (idToken) params.set('id_token_hint', idToken)
    if (POST_LOGOUT_URI) params.set('post_logout_redirect_uri', POST_LOGOUT_URI)
    window.location.assign(
      `${AUTH_BASE_URL}/v1/public/browser/logout${params.size ? `?${params}` : ''}`,
    )
  },
}
