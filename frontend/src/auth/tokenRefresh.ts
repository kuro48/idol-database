import { useAuthStore } from './authStore'
import { userManager } from './oidcClient'

const REFRESH_SKEW_MS = 60_000

let refreshPromise: Promise<void> | null = null

function tokenNeedsRefresh(expiresAt: number | null) {
  return expiresAt !== null && Date.now() + REFRESH_SKEW_MS >= expiresAt
}

async function refreshTokens() {
  const { refreshToken } = useAuthStore.getState()
  if (!refreshToken) return

  const refreshed = await userManager.refresh(refreshToken)
  useAuthStore.setState({
    accessToken: refreshed.access_token,
    idToken: refreshed.id_token ?? useAuthStore.getState().idToken,
    refreshToken: refreshed.refresh_token ?? refreshToken,
    tokenExpiresAt: Date.now() + refreshed.expires_in * 1000,
  })
}

export async function getValidAuthHeaders() {
  const state = useAuthStore.getState()
  if (state.refreshToken && tokenNeedsRefresh(state.tokenExpiresAt)) {
    refreshPromise ??= refreshTokens().finally(() => {
      refreshPromise = null
    })
    await refreshPromise
  }

  const { accessToken, idToken } = useAuthStore.getState()
  return { accessToken, idToken }
}
