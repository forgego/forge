import { useEffect } from "react"
import { useThemeStore } from "@/store/themeStore"
import { primaries } from "@/lib/themes"

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const { theme, primary, radius } = useThemeStore()

  // Handle Dark Mode
  useEffect(() => {
    const root = window.document.documentElement
    root.classList.remove("light", "dark")
    root.classList.add(theme)
  }, [theme])

  // Handle Primary Color and Radius
  useEffect(() => {
    const root = window.document.documentElement
    const primaryColor = primaries.find((p) => p.name === primary)

    if (primaryColor) {
      const cssVars = theme === "dark" ? primaryColor.cssVars.dark : primaryColor.cssVars.light
      Object.entries(cssVars).forEach(([key, value]) => {
        root.style.setProperty(key, value)
      })
    }
    
    root.style.setProperty("--radius", `${radius}rem`)
  }, [theme, primary, radius])

  return <>{children}</>
}
