import { cn } from "./utils"

/** Ordered categorical ramp. Index 0 is the iris accent. Both themes, via CSS vars. */
export const CHART_COLORS = [
  "hsl(var(--chart-1))",
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
  "hsl(var(--chart-6))",
] as const

/** Pick a series color by index, wrapping around the ramp. */
export function chartColor(index: number): string {
  return CHART_COLORS[index % CHART_COLORS.length]
}

/** Minimum height so axis labels are never cramped (VB-04). */
export const CHART_MIN_HEIGHT = 280

/** Shared <XAxis> / <YAxis> props. */
export const axisProps = {
  stroke: "hsl(var(--muted-foreground))",
  tick: { fill: "hsl(var(--muted-foreground))", fontSize: 12 },
  tickLine: false,
  axisLine: false,
} as const

/** Shared <CartesianGrid> props. */
export const gridProps = {
  stroke: "hsl(var(--grid-line))",
  strokeDasharray: "3 3",
  vertical: false,
} as const

export function ChartTooltip({ active, payload, label, className }: any) {
  if (!active || !payload?.length) return null
  return (
    <div
      className={cn(
        "rounded border border-border bg-surface-3 px-2 py-1.5 shadow-overlay",
        className
      )}
    >
      {label !== undefined && label !== null && (
        <p className="mb-1 text-micro text-muted-foreground">{String(label)}</p>
      )}
      {payload.map((entry: any, i: number) => (
        <div key={i} className="flex items-center gap-2 text-meta">
          <span
            aria-hidden
            className="h-2 w-2 shrink-0 rounded-full"
            style={{ background: entry.color ?? chartColor(i) }}
          />
          <span className="text-muted-foreground">{entry.name}</span>
          <span className="ml-auto font-mono tabular-nums text-foreground">
            {String(entry.value)}
          </span>
        </div>
      ))}
    </div>
  )
}