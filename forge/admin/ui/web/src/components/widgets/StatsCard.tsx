import { Card, CardContent } from "../ui/card";
import { cn } from "../../lib/utils";
import { TrendingUp, TrendingDown } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { ResponsiveContainer, AreaChart, Area } from "recharts";

interface StatsCardProps {
  title: string;
  value: string | number;
  description?: string;
  trend?: number; // percentage
  icon: LucideIcon;
  color?: "primary" | "success" | "warning" | "destructive";
  chartData?: any[];
}

export function StatsCard({
  title,
  value,
  description,
  trend,
  icon: Icon,
  color = "primary",
  chartData,
}: StatsCardProps) {
  const colorMap = {
    primary: "text-primary bg-primary/10",
    success: "text-emerald-500 bg-emerald-500/10",
    warning: "text-amber-500 bg-amber-500/10",
    destructive: "text-destructive bg-destructive/10",
  };

  const chartColorMap = {
    primary: "hsl(var(--primary))",
    success: "#10b981",
    warning: "#f59e0b",
    destructive: "#ef4444",
  };

  return (
    <Card className="glass-lite border-border/50 shadow-sm overflow-hidden group hover:shadow-md transition-all duration-300">
      <CardContent className="p-6">
        <div className="flex items-center justify-between mb-4">
          <div className={cn("p-2 rounded-lg transition-transform group-hover:scale-110 duration-300", colorMap[color])}>
            <Icon className="h-5 w-5" />
          </div>
          {trend !== undefined && (
            <div className={cn(
              "flex items-center gap-1 text-xs font-bold px-2 py-0.5 rounded-full",
              trend >= 0 ? "text-emerald-500 bg-emerald-500/10" : "text-destructive bg-destructive/10"
            )}>
              {trend >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
              {Math.abs(trend)}%
            </div>
          )}
        </div>
        
        <div className="space-y-1">
          <h3 className="text-xs font-bold uppercase tracking-widest text-muted-foreground/80">
            {title}
          </h3>
          <div className="text-3xl font-black tracking-tight text-foreground/90 tabular-nums">
            {value}
          </div>
          {description && (
            <p className="text-[11px] text-muted-foreground font-medium">
              {description}
            </p>
          )}
        </div>

        {chartData && (
          <div className="h-12 w-full mt-4 -mx-6 -mb-6 opacity-50 group-hover:opacity-100 transition-opacity duration-500">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id={`gradient-${title}`} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartColorMap[color]} stopOpacity={0.3} />
                    <stop offset="95%" stopColor={chartColorMap[color]} stopOpacity={0} />
                  </linearGradient>
                </defs>
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke={chartColorMap[color]}
                  strokeWidth={2}
                  fillOpacity={1}
                  fill={`url(#gradient-${title})`}
                  isAnimationActive={true}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
