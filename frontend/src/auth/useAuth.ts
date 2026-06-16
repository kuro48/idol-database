import { useAuthStore } from './authStore'
import type { OshiEntry } from './authStore'

export type { OshiEntry }

export function useAuth() {
  const accessToken = useAuthStore((s) => s.accessToken)
  const idToken = useAuthStore((s) => s.idToken)
  const refreshToken = useAuthStore((s) => s.refreshToken)
  const tokenExpiresAt = useAuthStore((s) => s.tokenExpiresAt)
  const email = useAuthStore((s) => s.email)
  const displayName = useAuthStore((s) => s.displayName)
  const oshiColor = useAuthStore((s) => s.oshiColor)
  const oshis = useAuthStore((s) => s.oshis)
  const canWrite = useAuthStore((s) => s.canWrite)
  const isAdmin = useAuthStore((s) => s.isAdmin)
  const setAuth = useAuthStore((s) => s.setAuth)
  const setOshiColor = useAuthStore((s) => s.setOshiColor)
  const setOshis = useAuthStore((s) => s.setOshis)
  const logout = useAuthStore((s) => s.logout)

  return {
    accessToken,
    idToken,
    refreshToken,
    tokenExpiresAt,
    email,
    displayName,
    oshiColor,
    oshis,
    canWrite,
    isAdmin,
    isLoggedIn: accessToken !== null,
    setAuth,
    setOshiColor,
    setOshis,
    logout,
  }
}
