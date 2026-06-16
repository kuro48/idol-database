import { create } from 'zustand'
import { createJSONStorage, persist } from 'zustand/middleware'

export interface OshiEntry {
  idol_id: string
  fan_since?: string
}

interface AuthState {
  accessToken: string | null
  idToken: string | null
  refreshToken: string | null
  tokenExpiresAt: number | null
  email: string | null
  displayName: string | null
  oshiColor: string | null
  oshis: OshiEntry[]
  canWrite: boolean
  isAdmin: boolean
  setAuth: (
    token: string,
    idToken: string | null,
    refreshToken: string | null,
    tokenExpiresAt: number | null,
    email: string,
    displayName: string,
    oshiColor: string,
    canWrite: boolean,
    isAdmin: boolean,
  ) => void
  setOshiColor: (color: string) => void
  setOshis: (oshis: OshiEntry[]) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      idToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
      email: null,
      displayName: null,
      oshiColor: null,
      oshis: [],
      canWrite: false,
      isAdmin: false,
      setAuth: (
        token,
        idToken,
        refreshToken,
        tokenExpiresAt,
        email,
        displayName,
        oshiColor,
        canWrite,
        isAdmin,
      ) =>
        set({
          accessToken: token,
          idToken,
          refreshToken,
          tokenExpiresAt,
          email,
          displayName,
          oshiColor,
          canWrite,
          isAdmin,
        }),
      setOshiColor: (color) => set({ oshiColor: color }),
      setOshis: (oshis) => set({ oshis }),
      logout: () =>
        set({
          accessToken: null,
          idToken: null,
          refreshToken: null,
          tokenExpiresAt: null,
          email: null,
          displayName: null,
          oshiColor: null,
          oshis: [],
          canWrite: false,
          isAdmin: false,
        }),
    }),
    {
      name: 'idol-db-auth',
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        accessToken: state.accessToken,
        idToken: state.idToken,
        refreshToken: state.refreshToken,
        tokenExpiresAt: state.tokenExpiresAt,
        email: state.email,
        displayName: state.displayName,
        oshiColor: state.oshiColor,
        oshis: state.oshis,
        canWrite: state.canWrite,
        isAdmin: state.isAdmin,
      }),
      merge: (persisted, current) => {
        const saved = persisted as Partial<AuthState>
        return {
          ...current,
          accessToken: saved.accessToken ?? null,
          idToken: saved.idToken ?? null,
          refreshToken: saved.refreshToken ?? null,
          tokenExpiresAt: saved.tokenExpiresAt ?? null,
          email: saved.email ?? null,
          displayName: saved.displayName ?? null,
          oshiColor: saved.oshiColor ?? null,
          oshis: saved.oshis ?? [],
          canWrite: saved.canWrite ?? false,
          isAdmin: saved.isAdmin ?? false,
        }
      },
    },
  ),
)
