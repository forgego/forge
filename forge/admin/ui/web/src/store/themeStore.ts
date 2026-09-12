import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type Theme = "light" | "dark" | "system"

export const DEFAULT_THEME = {
  theme: "system" as Theme,
  primary: "iris",
  radius: 0.625,
}

interface ThemeState {
  theme: Theme
  primary: string
  radius: number
  /** Resolved light/dark value after applying the "system" preference. */
  resolvedTheme: "light" | "dark"
  setTheme: (theme: Theme) => void
  setPrimary: (primary: string) => void
  setRadius: (radius: number) => void
  resetTheme: () => void
}

function systemTheme(): "light" | "dark" {
  if (typeof window !== "undefined" && window.matchMedia?.("(prefers-color-scheme: dark)").matches) {
    return "dark"
  }
  return "light"
}

function resolveTheme(theme: Theme): "light" | "dark" {
  return theme === "system" ? systemTheme() : theme
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      ...DEFAULT_THEME,
      resolvedTheme: resolveTheme(DEFAULT_THEME.theme),
      setTheme: (theme) => set({ theme, resolvedTheme: resolveTheme(theme) }),
      setPrimary: (primary) => set({ primary }),
      setRadius: (radius) => set({ radius }),
      resetTheme: () =>
        set({
          ...DEFAULT_THEME,
          resolvedTheme: resolveTheme(DEFAULT_THEME.theme),
        }),
    }),
    {
      name: "admin-theme-storage",
      onRehydrateStorage: () => (state) => {
        // Re-resolve "system" on load: the OS preference may have changed.
        if (state && state.theme === "system") {
          state.resolvedTheme = systemTheme()
        }
      },
    }
  )
)

// Keep resolvedTheme in sync when the OS preference flips while "system" is active.
if (typeof window !== "undefined" && window.matchMedia) {
  window
    .matchMedia("(prefers-color-scheme: dark)")
    .addEventListener?.("change", () => {
      const { theme } = useThemeStore.getState()
      if (theme === "system") {
        useThemeStore.setState({ resolvedTheme: systemTheme() })
      }
    })
}
