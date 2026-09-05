import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

interface ChartWidgetProps {
  title: string;
  data: { [key: string]: any }[];
  dataKey?: string;
  xAxisKey?: string;
  lines?: { key: string; color: string; name: string }[];
  areas?: { key: string; color: string; name: string }[];
  bars?: { key: string; color: string; name: string }[];
  type?: 'line' | 'area' | 'bar';
  height?: number;
  showLegend?: boolean;
  showGrid?: boolean;
  format?: 'currency' | 'number' | 'percent' | 'none';
}

export function ChartWidget({
  title,
  data,
  dataKey = 'value',
  xAxisKey = 'name',
  lines,
  areas,
  bars,
  type = 'line',
  height = 300,
  showLegend = true,
  showGrid = true,
  format = 'none',
}: ChartWidgetProps) {
  const formatValue = (value: number) => {
    switch (format) {
      case 'currency':
        return new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'USD',
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(value);
      case 'number':
        return new Intl.NumberFormat('en-US').format(value);
      case 'percent':
        return `${value.toFixed(1)}%`;
      default:
        return value;
    }
  };

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-popover border border-border rounded-lg shadow-lg p-3">
          <p className="text-sm font-medium mb-2">{label}</p>
          {payload.map((entry: any, idx: number) => (
            <p key={idx} className="text-sm" style={{ color: entry.color }}>
              {entry.name}: {formatValue(entry.value)}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  const renderChart = () => {
    const chartProps = {
      data,
      margin: { top: 10, right: 30, left: 0, bottom: 0 },
    };

    switch (type) {
      case 'area':
        return (
          <AreaChart {...chartProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey={xAxisKey}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
            />
            <YAxis
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
              tickFormatter={format !== 'none' ? formatValue : undefined}
            />
            <Tooltip content={<CustomTooltip />} />
            {showLegend && <Legend />}
            {areas?.map((area, idx) => (
              <Area
                key={idx}
                type="monotone"
                dataKey={area.key}
                stroke={area.color}
                fill={area.color}
                fillOpacity={0.3}
                name={area.name}
              />
            ))}
          </AreaChart>
        );

      case 'bar':
        return (
          <BarChart {...chartProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey={xAxisKey}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
            />
            <YAxis
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
              tickFormatter={format !== 'none' ? formatValue : undefined}
            />
            <Tooltip content={<CustomTooltip />} />
            {showLegend && <Legend />}
            {bars?.map((bar, idx) => (
              <Bar key={idx} dataKey={bar.key} fill={bar.color} name={bar.name} radius={[4, 4, 0, 0]} />
            ))}
          </BarChart>
        );

      case 'line':
      default:
        return (
          <LineChart {...chartProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey={xAxisKey}
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
            />
            <YAxis
              className="text-xs"
              tick={{ fill: 'hsl(var(--muted-foreground))' }}
              axisLine={{ stroke: 'hsl(var(--border))' }}
              tickFormatter={format !== 'none' ? formatValue : undefined}
            />
            <Tooltip content={<CustomTooltip />} />
            {showLegend && <Legend />}
            {lines?.map((line, idx) => (
              <Line
                key={idx}
                type="monotone"
                dataKey={line.key}
                stroke={line.color}
                strokeWidth={2}
                dot={{ fill: line.color, strokeWidth: 2, r: 4 }}
                activeDot={{ r: 6, strokeWidth: 2 }}
                name={line.name}
              />
            ))}
          </LineChart>
        );
    }
  };

  return (
    <Card className="glass-lite border-border/50 shadow-sm">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-bold">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ height }}>
          <ResponsiveContainer width="100%" height="100%">
            {renderChart()}
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}

interface SalesChartProps {
  data?: { date: string; sales: number; orders: number }[];
  height?: number;
}

export function SalesChart({ data, height = 300 }: SalesChartProps) {
  const chartData = data || [
    { date: 'Mon', sales: 4000, orders: 24 },
    { date: 'Tue', sales: 3000, orders: 18 },
    { date: 'Wed', sales: 5000, orders: 32 },
    { date: 'Thu', sales: 4500, orders: 28 },
    { date: 'Fri', sales: 6000, orders: 42 },
    { date: 'Sat', sales: 5500, orders: 38 },
    { date: 'Sun', sales: 4800, orders: 31 },
  ];

  return (
    <ChartWidget
      title="Sales & Orders"
      data={chartData}
      xAxisKey="date"
      lines={[
        { key: 'sales', color: 'hsl(var(--primary))', name: 'Sales ($)' },
        { key: 'orders', color: 'hsl(142 76% 36%)', name: 'Orders' },
      ]}
      type="line"
      height={height}
      format="currency"
    />
  );
}

export function RevenueChart({ data, height = 300 }: SalesChartProps) {
  const chartData = data || [
    { date: 'Jan', revenue: 125000 },
    { date: 'Feb', revenue: 132000 },
    { date: 'Mar', revenue: 141000 },
    { date: 'Apr', revenue: 138000 },
    { date: 'May', revenue: 155000 },
    { date: 'Jun', revenue: 172000 },
    { date: 'Jul', revenue: 168000 },
  ];

  return (
    <ChartWidget
      title="Revenue Trend"
      data={chartData}
      xAxisKey="date"
      areas={[{ key: 'revenue', color: 'hsl(var(--primary))', name: 'Revenue' }]}
      type="area"
      height={height}
      format="currency"
    />
  );
}

export function TrafficChart({ data, height = 300 }: SalesChartProps) {
  const chartData = data || [
    { date: 'Mon', visitors: 2400, pageViews: 8400 },
    { date: 'Tue', visitors: 2100, pageViews: 7200 },
    { date: 'Wed', visitors: 2800, pageViews: 9600 },
    { date: 'Thu', visitors: 3200, pageViews: 11200 },
    { date: 'Fri', visitors: 3600, pageViews: 12800 },
    { date: 'Sat', visitors: 2900, pageViews: 9800 },
    { date: 'Sun', visitors: 2500, pageViews: 8600 },
  ];

  return (
    <ChartWidget
      title="Traffic Overview"
      data={chartData}
      xAxisKey="date"
      bars={[
        { key: 'visitors', color: 'hsl(217 91% 60%)', name: 'Visitors' },
        { key: 'pageViews', color: 'hsl(142 76% 36%)', name: 'Page Views' },
      ]}
      type="bar"
      height={height}
    />
  );
}
