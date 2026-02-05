import { widgetRegistry, WIDGET_TYPES } from './widgets';
import StatsWidget from '../features/dashboard/widgets/StatsWidget';
import ChartWidget from '../features/dashboard/widgets/ChartWidget';
import ActivityWidget from '../features/dashboard/widgets/ActivityWidget';

export function bootstrapAdmin() {
  widgetRegistry.register(WIDGET_TYPES.STATS, StatsWidget);
  widgetRegistry.register(WIDGET_TYPES.CHART, ChartWidget);
  widgetRegistry.register(WIDGET_TYPES.ACTIVITY, ActivityWidget);
}
