import { create } from 'zustand'
import { createJSONStorage, persist } from 'zustand/middleware'

interface AuthState {
  accessToken: string | null
  idToken: string | null
  refreshToken: string | null
  email: string | null
  displayName: string | null
  oshiColor: string | null
  canWrite: boolean
  isAdmin: boolean
  setAuth: (
    token: string,
    idToken: string | null,
    refreshToken: string | null,
    email: string,
    displayName: string,
    oshiColor: string,
    canWrite: boolean,
    isAdmin: boolean,
  ) => void
  setOshiColor: (color: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      idToken: null,
      refreshToken: null,
      email: null,
      displayName: null,
      oshiColor: null,
      canWrite: false,
      isAdmin: false,
      setAuth: (token, idToken, refreshToken, email, displayName, oshiColor, canWrite, isAdmin) =>
        set({
          accessToken: token,
          idToken,
          refreshToken,
          email,
          displayName,
          oshiColor,
          canWrite,
          isAdmin,
        }),
      setOshiColor: (color) => set({ oshiColor: color }),
      logout: () =>
        set({
          accessToken: null,
          idToken: null,
          refreshToken: null,
          email: null,
          displayName: null,
          oshiColor: null,
          canWrite: false,
          isAdmin: false,
        }),
    }),
    {
      name: 'idol-db-auth',
      storage: createJSONStorage(() => sessionStorage),
    },
  ),
)
