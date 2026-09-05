export type Theme = "light" | "dark" | "system"

export const info = {
  name: "Forge Admin",
  version: "1.0.0",
}

export const primaries = [
  {
    name: "default",
    label: "Default (Blue)",
    active: "221 83% 53%",
    cssVars: {
      light: {
        "--primary": "221 83% 53%",
        "--primary-foreground": "210 40% 98%",
        "--ring": "221 83% 53%",
      },
      dark: {
        "--primary": "217 91% 60%",
        "--primary-foreground": "222 47% 11%",
        "--ring": "217 91% 60%",
      },
    },
  },
  {
    name: "violet",
    label: "Violet",
    active: "262.1 83.3% 57.8%",
    cssVars: {
      light: {
        "--primary": "262.1 83.3% 57.8%",
        "--primary-foreground": "210 40% 98%",
        "--ring": "262.1 83.3% 57.8%",
      },
      dark: {
        "--primary": "263.4 70% 50.4%",
        "--primary-foreground": "210 40% 98%",
        "--ring": "263.4 70% 50.4%",
      },
    },
  },
  {
    name: "green",
    label: "Green",
    active: "142.1 76.2% 36.3%",
    cssVars: {
      light: {
        "--primary": "142.1 76.2% 36.3%",
        "--primary-foreground": "355.7 100% 97.3%",
        "--ring": "142.1 76.2% 36.3%",
      },
      dark: {
        "--primary": "142.1 70.6% 45.3%",
        "--primary-foreground": "144.9 80.4% 10%",
        "--ring": "142.1 70.6% 45.3%",
      },
    },
  },
  {
    name: "red",
    label: "Red",
    active: "346.8 77.2% 49.8%",
    cssVars: {
      light: {
        "--primary": "346.8 77.2% 49.8%",
        "--primary-foreground": "355.7 100% 97.3%",
        "--ring": "346.8 77.2% 49.8%",
      },
      dark: {
        "--primary": "346.8 77.2% 49.8%",
        "--primary-foreground": "355.7 100% 97.3%",
        "--ring": "346.8 77.2% 49.8%",
      },
    },
  },
  {
    name: "orange",
    label: "Orange",
    active: "24.6 95% 53.1%",
    cssVars: {
      light: {
        "--primary": "24.6 95% 53.1%",
        "--primary-foreground": "60 9.1% 97.8%",
        "--ring": "24.6 95% 53.1%",
      },
      dark: {
        "--primary": "20.5 90.2% 48.2%",
        "--primary-foreground": "60 9.1% 97.8%",
        "--ring": "20.5 90.2% 48.2%",
      },
    },
  },
]
