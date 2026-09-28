import { componentRegistry } from '../lib/registry';
import React from 'react';

/**
 * Hook to resolve an overridable UI component.
 * If a custom implementation is registered in the componentRegistry, it returns that.
 * Otherwise, it returns the provided default component.
 */
export function useUIComponent<T = any>(
  name: string,
  defaultComponent: React.ElementType<T>
): React.ElementType<T> {
  const customComponent = componentRegistry.get<T>(name);
  return (customComponent as React.ElementType<T>) || defaultComponent;
}
