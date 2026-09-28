import React from 'react';

export interface UIComponentProps {
  [key: string]: any;
}

class UIComponentRegistry {
  private components: Map<string, React.ComponentType<any>> = new Map();

  /**
   * Registers a UI component that can be used or overridden globally.
   */
  register(name: string, component: React.ComponentType<any>) {
    this.components.set(name, component);
  }

  /**
   * Retrieves a registered component by its unique name.
   */
  get<T = any>(name: string): React.ComponentType<T> | undefined {
    return this.components.get(name) as React.ComponentType<T> | undefined;
  }

  registerAll(components: Record<string, React.ComponentType<any>>) {
    Object.entries(components).forEach(([name, component]) => {
      this.register(name, component);
    });
  }
}

export const componentRegistry = new UIComponentRegistry();

// Standard unique names for core admin components
export const CORE_COMPONENTS = {
  PAGES: {
    LIST: 'forge.pages.list',
    DETAIL: 'forge.pages.detail',
    FORM: 'forge.pages.form',
    DASHBOARD: 'forge.pages.dashboard',
  },
  LAYOUT: {
    SIDEBAR: 'forge.layout.sidebar',
    HEADER: 'forge.layout.header',
    TABS: 'forge.layout.tabs',
  },
  WIDGETS: {
    BASE: 'forge.widgets.base',
  },
} as const;
