import React, { Suspense } from 'react';
import { widgetRegistry, WIDGET_TYPES } from './widgets';
import type { WidgetProps } from './widgets';
import StatsWidget from '../features/dashboard/widgets/StatsWidget';
import ActivityWidget from '../features/dashboard/widgets/ActivityWidget';
import { Skeleton } from '../components/ui/skeleton';

// recharts is heavy and only needed by the chart widget, so load it lazily to
// keep it out of the initial bundle. The wrapper renders its own Suspense
// boundary so registry consumers do not need one.
const LazyDashboardChartWidget = React.lazy(
  () => import('../features/dashboard/widgets/ChartWidget'),
);

function DashboardChartWidget(props: WidgetProps) {
  return React.createElement(
    Suspense,
    { fallback: React.createElement(Skeleton, { className: 'h-[280px] w-full' }) },
    React.createElement(LazyDashboardChartWidget, props),
  );
}

export function bootstrapAdmin() {
  widgetRegistry.register(WIDGET_TYPES.STATS, StatsWidget);
  widgetRegistry.register(WIDGET_TYPES.CHART, DashboardChartWidget);
  widgetRegistry.register(WIDGET_TYPES.ACTIVITY, ActivityWidget);
}
