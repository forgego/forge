import { cn } from "@/lib/utils";
import { StatusBadge } from "@/components/ui/status-badge";
import { forwardRef, Suspense, lazy } from "react";
import { TrendingDown, TrendingUp } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

// The sparkline pulls in recharts, so it is loaded lazily: StatTile instances
// without chartData (the common case) never download the charts code.
const StatSparkline = lazy(() =>
  import("./stat-sparkline").then((m) => ({ default: m.StatSparkline }))
);

export interface StatTileProps extends React.HTMLAttributes<HTMLElement> {
  label: string
  value: string | number
  description?: string
  /** Signed percentage change. Positive renders success, negative renders danger. */
  delta?: number
  icon?: LucideIcon
  /** Recharts rows for the sparkline. */
  chartData?: Array<Record<string, unknown>>
  /** Key in chartData to plot. Defaults to "value". */
  chartKey?: string
}

export const StatTile = forwardRef<
  HTMLElement,
  StatTileProps
>(({ 
  label, 
  value, 
  description, 
  delta, 
  icon: Icon,
  chartData,
  chartKey = "value",
  className,
  ...props
}, ref) => {
  return (
    <article
      ref={ref}
      className={cn(
        "group flex flex-col gap-canvas rounded-lg border border-border-subtle bg-surface-2 p-canvas transition-colors duration-base ease-out hover:border-border",
        className
      )}
      {...props}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-micro text-muted-foreground">{label}</p>
        {Icon && (
          <span className="text-muted-foreground/60">
            <Icon className="h-4 w-4" aria-hidden />
          </span>
        )}
      </div>
      
      <div className="flex items-end justify-between gap-3">
        <p className="font-mono text-metric tabular-nums text-foreground">
          {value}
        </p>
        {delta !== undefined && (
          <StatusBadge tone={delta >= 0 ? "success" : "danger"}>
            {delta >= 0 ? <TrendingUp className="h-3 w-3" aria-hidden /> : <TrendingDown className="h-3 w-3" aria-hidden />}
            {Math.abs(delta)}%
          </StatusBadge>
        )}
      </div>
      
      {chartData && chartData.length > 0 && (
        <Suspense fallback={<Skeleton className="h-10 w-full" />}>
          <StatSparkline data={chartData} dataKey={chartKey} />
        </Suspense>
      )}
      
      {description && (
        <p className="text-meta text-muted-foreground">{description}</p>
      )}
    </article>
  );
});

StatTile.displayName = "StatTile";