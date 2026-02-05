import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type Theme = "light" | "dark"

interface ThemeState {
  theme: Theme
  primary: string
  radius: number
  setTheme: (theme: Theme) => void
  setPrimary: (primary: string) => void
  setRadius: (radius: number) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: "light",
      primary: "default",
      radius: 0.5,
      setTheme: (theme) => set({ theme }),
      setPrimary: (primary) => set({ primary }),
      setRadius: (radius) => set({ radius }),
    }),
    {
      name: "admin-theme-storage",
    }
  )
)
