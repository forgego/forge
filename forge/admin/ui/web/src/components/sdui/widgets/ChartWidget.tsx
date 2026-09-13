import React from 'react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  AreaChart,
  Area,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '../../ui/card';
import { CHART_COLORS, axisProps, gridProps, ChartTooltip, CHART_MIN_HEIGHT } from '../../../lib/chart-theme';

interface ChartWidgetProps {
  title?: string;
  chartType?: 'line' | 'bar' | 'area';
  data: any[];
  dataKeys?: string[]; // Keys to plot (e.g. ["sales", "sales", "profit")
  xAxisKey?: string;   // Key for X Axis (e.g. "date")
  height?: number;
  colors?: string[];
}

export const ChartWidget: React.FC<ChartWidgetProps> = ({
  title,
  chartType = 'line',
  data,
  dataKeys,
  xAxisKey = 'name',
  height = 300,
  colors = CHART_COLORS,
}) => {
  // Infer data keys if not provided (exclude xAxisKey)
  const keys = dataKeys || (data.length > 0 ? Object.keys(data[0]).filter(k => k !== xAxisKey) : []);

  const renderChart = () => {
    const commonProps = {
      data,
      margin: { top: 5, right: 30, left: 20, bottom: 5 },
    };

    switch (chartType) {
case 'bar':
         return (
           <BarChart {...commonProps}>
             <CartesianGrid {...gridProps} />
             <XAxis dataKey={xAxisKey} {...axisProps} />
             <YAxis {...axisProps} />
             <Tooltip 
               content={<ChartTooltip />}
               cursor={{ fill: "hsl(var(--surface-sunken))" }}
             />
             <Legend />
             {keys.map((key, index) => (
               <Bar 
                 key={key} 
                 dataKey={key} 
                 fill={colors[index % colors.length]} 
                 radius={[4, 4, 0, 0]} 
               />
             ))}
           </BarChart>
         );
case 'area':
         return (
           <AreaChart {...commonProps}>
             <CartesianGrid {...gridProps} />
             <XAxis dataKey={xAxisKey} {...axisProps} />
             <YAxis {...axisProps} />
             <Tooltip 
               content={<ChartTooltip />}
               cursor={{ fill: "hsl(var(--surface-sunken))" }}
             />
             <Legend />
             {keys.map((key, index) => (
               <Area 
                 key={key} 
                 type="monotone" 
                 dataKey={key} 
                 stroke={colors[index % colors.length]} 
                 fill={colors[index % colors.length]} 
                 fillOpacity={0.2} 
               />
             ))}
           </AreaChart>
         );
default:
         return (
           <LineChart {...commonProps}>
             <CartesianGrid {...gridProps} />
             <XAxis dataKey={xAxisKey} {...axisProps} />
             <YAxis {...axisProps} />
             <Tooltip 
               content={<ChartTooltip />}
               cursor={{ fill: "hsl(var(--surface-sunken))" }}
             />
             <Legend />
             {keys.map((key, index) => (
               <Line 
                 key={key} 
                 type="monotone" 
                 dataKey={key} 
                 stroke={colors[index % colors.length]} 
                 strokeWidth={2}
                 dot={{ r: 4 }}
                 activeDot={{ r: 6 }}
               />
             ))}
           </LineChart>
         );
    }
  };

return (
     <Card className="col-span-full">
       {title && (
         <CardHeader>
           <CardTitle>{title}</CardTitle>
         </CardHeader>
       )}
       <CardContent>
         <div style={{ width: '100%', height: Math.max(height, CHART_MIN_HEIGHT) }}>
           <ResponsiveContainer>
             {renderChart()}
           </ResponsiveContainer>
         </div>
       </CardContent>
     </Card>
   );
};
