/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        "surface-1": "hsl(var(--surface-1))",
        "surface-2": "hsl(var(--surface-2))",
        "surface-3": "hsl(var(--surface-3))",
        "surface-sunken": "hsl(var(--surface-sunken))",
        "border-subtle": "hsl(var(--border-subtle))",
        "border-strong": "hsl(var(--border-strong))",
        "grid-line": "hsl(var(--grid-line))",
        success: {
          DEFAULT: "hsl(var(--success))",
          fg: "hsl(var(--success-fg))",
          surface: "hsl(var(--success-surface))",
        },
        warning: {
          DEFAULT: "hsl(var(--warning))",
          fg: "hsl(var(--warning-fg))",
          surface: "hsl(var(--warning-surface))",
        },
        danger: {
          DEFAULT: "hsl(var(--danger))",
          fg: "hsl(var(--danger-fg))",
          surface: "hsl(var(--danger-surface))",
        },
        info: {
          DEFAULT: "hsl(var(--info))",
          fg: "hsl(var(--info-fg))",
          surface: "hsl(var(--info-surface))",
        },
        neutral: {
          DEFAULT: "hsl(var(--neutral))",
          fg: "hsl(var(--neutral-fg))",
          surface: "hsl(var(--neutral-surface))",
        },
        chart: {
          1: "hsl(var(--chart-1))",
          2: "hsl(var(--chart-2))",
          3: "hsl(var(--chart-3))",
          4: "hsl(var(--chart-4))",
          5: "hsl(var(--chart-5))",
          6: "hsl(var(--chart-6))",
        },
      },
      fontFamily: {
        sans: ["var(--font-sans)"],
        mono: ["var(--font-mono)"],
      },
      fontSize: {
        micro: ["var(--text-micro)", { lineHeight: "1.4", letterSpacing: "0.08em", fontWeight: "600" }],
        meta: ["var(--text-meta)", { lineHeight: "1.45", letterSpacing: "0" }],
        body: ["var(--text-body)", { lineHeight: "1.45", letterSpacing: "0" }],
        ui: ["var(--text-ui)", { lineHeight: "1.45", letterSpacing: "0" }],
        lead: ["var(--text-lead)", { lineHeight: "1.35", letterSpacing: "-0.01em", fontWeight: "600" }],
        title: ["var(--text-title)", { lineHeight: "1.2", letterSpacing: "-0.02em", fontWeight: "600" }],
        display: ["var(--text-display)", { lineHeight: "1.2", letterSpacing: "-0.02em", fontWeight: "600" }],
        metric: ["var(--text-metric)", { lineHeight: "1", letterSpacing: "-0.03em", fontWeight: "700" }],
        "metric-xl": ["var(--text-metric-xl)", { lineHeight: "1", letterSpacing: "-0.03em", fontWeight: "700" }],
      },
      spacing: {
        command: "var(--row-command)",
        grid: "var(--row-grid)",
        "grid-comfy": "var(--row-grid-comfy)",
        "pad-command": "var(--pad-command)",
        "pad-grid": "var(--pad-grid)",
        canvas: "var(--pad-canvas)",
        section: "var(--space-section)",
        gutter: "var(--page-gutter)",
      },
      borderRadius: {
        sm: "var(--radius-sm)",
        DEFAULT: "var(--radius)",
        md: "var(--radius)",
        lg: "var(--radius-lg)",
      },
      boxShadow: {
        overlay: "var(--elev-overlay)",
        dialog: "var(--elev-dialog)",
      },
      transitionTimingFunction: {
        out: "var(--ease-out)",
        "in-out": "var(--ease-in-out)",
      },
      transitionDuration: {
        fast: "var(--dur-fast)",
        base: "var(--dur-base)",
        slow: "var(--dur-slow)",
      },
      maxWidth: {
        page: "var(--page-max)",
      },
    },
  },
  plugins: [],
}