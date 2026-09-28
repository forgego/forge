import { ResponsiveContainer, AreaChart, Area } from "recharts";

export interface StatSparklineProps {
  /** Recharts rows to plot. */
  data: Array<Record<string, unknown>>;
  /** Key in data to plot. Defaults to "value". */
  dataKey?: string;
}

export function StatSparkline({ data, dataKey = "value" }: StatSparklineProps) {
  return (
    <div className="-mx-1 h-10">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
          <Area
            type="monotone"
            dataKey={dataKey}
            stroke="hsl(var(--chart-1))"
            fill="hsl(var(--chart-1))"
            fillOpacity={0.12}
            strokeWidth={1.5}
            isAnimationActive={false}
            dot={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
