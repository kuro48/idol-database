import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  accessToken: string | null
  email: string | null
  displayName: string | null
  oshiColor: string | null
  canWrite: boolean
  isAdmin: boolean
  setAuth: (
    token: string,
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
      email: null,
      displayName: null,
      oshiColor: null,
      canWrite: false,
      isAdmin: false,
      setAuth: (token, email, displayName, oshiColor, canWrite, isAdmin) =>
        set({
          accessToken: token,
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
          email: null,
          displayName: null,
          oshiColor: null,
          canWrite: false,
          isAdmin: false,
        }),
    }),
    { name: 'idol-db-auth' },
  ),
)
