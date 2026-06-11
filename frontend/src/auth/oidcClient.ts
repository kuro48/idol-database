interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  refresh_token?: string
  id_token?: string
  scope?: string
}

interface LoginTransaction {
  state: string
  codeVerifier: string
  clientId: string
  redirectUri: string
}

interface LoginRedirectOptions {
  returnTo?: string
}

const AUTH_BASE_URL = (
  import.meta.env.VITE_IDOL_AUTH_BASE_URL ??
  import.meta.env.VITE_OIDC_ISSUER ??
  ''
).replace(/\/$/, '')
const CLIENT_ID = import.meta.env.VITE_OIDC_CLIENT_ID
const REDIRECT_URI = import.meta.env.VITE_OIDC_REDIRECT_URI
const POST_LOGOUT_URI = import.meta.env.VITE_OIDC_POST_LOGOUT_URI
const SCOPE =
  import.meta.env.VITE_OIDC_SCOPE ?? 'openid email profile offline_access'
const TX_STORAGE_KEY = 'idol_auth_tx'
const RETURN_TO_STORAGE_KEY = 'idol-db-auth-return-to'

function storage() {
  if (!window.sessionStorage) {
    throw new Error('sessionStorage が利用できません。')
  }
  return window.sessionStorage
}

function randomBase64Url(bytes: number) {
  const values = new Uint8Array(bytes)
  crypto.getRandomValues(values)
  return base64Url(values)
}

function randomState() {
  return typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : randomBase64Url(24)
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

function normalizeReturnTo(returnTo?: string | null) {
  if (!returnTo || !returnTo.startsWith('/') || returnTo.startsWith('//')) {
    return '/idols'
  }
  if (returnTo.startsWith('/callback') || returnTo.startsWith('/login')) {
    return '/idols'
  }
  return returnTo
}

function browserLoginUrl(params: {
  clientId: string
  redirectUri: string
  scope: string
  state: string
  codeChallenge: string
  codeChallengeMethod: string
}) {
  const query = new URLSearchParams({
    client_id: params.clientId,
    redirect_uri: params.redirectUri,
    response_type: 'code',
    scope: params.scope,
    state: params.state,
    code_challenge: params.codeChallenge,
    code_challenge_method: params.codeChallengeMethod,
  })
  return `${AUTH_BASE_URL}/v1/public/browser/login?${query}`
}

async function postToken(body: URLSearchParams): Promise<TokenResponse> {
  const res = await fetch(`${AUTH_BASE_URL}/v1/public/api/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  })
  if (!res.ok) {
    throw new Error(`token request failed: ${res.status}`)
  }
  return (await res.json()) as TokenResponse
}

export const userManager = {
  async signinRedirect(options: LoginRedirectOptions = {}) {
    requireConfig()
    const state = randomState()
    const codeVerifier = randomBase64Url(48)
    const codeChallenge = await sha256Base64Url(codeVerifier)
    const tx: LoginTransaction = {
      state,
      codeVerifier,
      clientId: CLIENT_ID,
      redirectUri: REDIRECT_URI,
    }

    const store = storage()
    store.setItem(TX_STORAGE_KEY, JSON.stringify(tx))
    store.setItem(RETURN_TO_STORAGE_KEY, normalizeReturnTo(options.returnTo))

    window.location.assign(
      browserLoginUrl({
        clientId: CLIENT_ID,
        redirectUri: REDIRECT_URI,
        scope: SCOPE,
        state,
        codeChallenge,
        codeChallengeMethod: 'S256',
      }),
    )
  },

  async registrationRedirect(returnTo?: string) {
    requireConfig()
    const query = returnTo
      ? `?return_to=${encodeURIComponent(returnTo)}`
      : ''
    window.location.assign(`${AUTH_BASE_URL}/v1/public/browser/registration${query}`)
  },

  async signinRedirectCallback(): Promise<TokenResponse> {
    requireConfig()
    const params = new URL(window.location.href).searchParams
    const store = storage()
    const error = params.get('error')
    if (error) {
      store.removeItem(TX_STORAGE_KEY)
      throw new Error(params.get('error_description') ?? error)
    }

    const code = params.get('code')
    if (!code) {
      throw new Error('ログインコールバックに code が含まれていません。')
    }

    const saved = store.getItem(TX_STORAGE_KEY)
    if (!saved) {
      throw new Error('ログイン transaction が見つかりません。')
    }

    const tx = JSON.parse(saved) as LoginTransaction
    if (params.get('state') !== tx.state) {
      store.removeItem(TX_STORAGE_KEY)
      throw new Error('ログイン state が一致しません。')
    }
    store.removeItem(TX_STORAGE_KEY)

    const body = new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: tx.redirectUri,
      client_id: tx.clientId,
      code_verifier: tx.codeVerifier,
    })
    return postToken(body)
  },

  consumeReturnTo() {
    const store = storage()
    const returnTo = normalizeReturnTo(store.getItem(RETURN_TO_STORAGE_KEY))
    store.removeItem(RETURN_TO_STORAGE_KEY)
    return returnTo
  },

  async refresh(refreshToken: string): Promise<TokenResponse> {
    requireConfig()
    return postToken(
      new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: CLIENT_ID,
      }),
    )
  },

  async revoke(token: string) {
    requireConfig()
    const body = new URLSearchParams({
      token,
      client_id: CLIENT_ID,
    })
    const res = await fetch(`${AUTH_BASE_URL}/v1/public/api/token/revoke`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: body.toString(),
    })
    if (!res.ok) {
      throw new Error(`token revoke failed: ${res.status}`)
    }
  },

  async signoutRedirect(idToken?: string | null) {
    requireConfig()
    const params = new URLSearchParams()
    if (idToken) params.set('id_token_hint', idToken)
    if (POST_LOGOUT_URI) params.set('post_logout_redirect_uri', POST_LOGOUT_URI)
    params.set('state', randomState())
    window.location.assign(
      `${AUTH_BASE_URL}/v1/public/browser/logout${params.size ? `?${params}` : ''}`,
    )
  },
}

export type { TokenResponse }
