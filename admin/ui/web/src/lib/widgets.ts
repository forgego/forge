import React from 'react';

export interface WidgetProps {
  data: any;
  config: WidgetConfig;
}

export interface WidgetConfig {
  id: string;
  type: string;
  title: string;
  description?: string;
  size?: 'sm' | 'md' | 'lg' | 'full';
  params?: Record<string, any>;
}

class WidgetRegistry {
  private widgets: Map<string, React.ComponentType<WidgetProps>> = new Map();

  register(type: string, component: React.ComponentType<WidgetProps>) {
    this.widgets.set(type, component);
  }

  get(type: string): React.ComponentType<WidgetProps> | undefined {
    return this.widgets.get(type);
  }
}

export const widgetRegistry = new WidgetRegistry();

// Standard widget types
export const WIDGET_TYPES = {
  STATS: 'stats',
  CHART: 'chart',
  TABLE: 'table',
  ACTIVITY: 'activity',
  RECENT: 'recent',
  CUSTOM: 'custom',
} as const;
