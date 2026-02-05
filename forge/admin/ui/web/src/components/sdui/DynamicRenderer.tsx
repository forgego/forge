import React from "react";
import { Card, CardHeader, CardTitle, CardContent } from "../ui/card";
import { Button } from "../ui/button";
import { cn } from "../../lib/utils";
import { ChartWidget } from "./widgets/ChartWidget";

// Component Types matching Go definitions
export type ComponentType =
  | "page"
  | "card"
  | "grid"
  | "text"
  | "button"
  | "stats"
  | "chart"
  | "container";

export interface LayoutConfig {
  col_span?: number;
  row_span?: number;
  align?: string;
  justify?: string;
  padding?: string;
  margin?: string;
  class_name?: string;
}

export interface ComponentData {
  type: ComponentType;
  id?: string;
  props?: Record<string, any>;
  children?: ComponentData[];
  layout?: LayoutConfig;
}

// Registry Context
const ComponentRegistryContext = React.createContext<
  Record<string, React.ComponentType<any>>
>({});

export const useComponentRegistry = () =>
  React.useContext(ComponentRegistryContext);

// Dynamic Renderer
export const DynamicRenderer: React.FC<{ component: ComponentData }> = ({
  component,
}) => {
  const registry = useComponentRegistry();
  const Component = registry[component.type];

  if (!Component) {
    console.warn(`Unknown component type: ${component.type}`);
    return null;
  }

  // Apply layout styles
  const layoutStyle: React.CSSProperties = {};
  const layoutClasses = [];

  if (component.layout) {
    if (component.layout.col_span)
      layoutClasses.push(`col-span-${component.layout.col_span}`);
    if (component.layout.row_span)
      layoutClasses.push(`row-span-${component.layout.row_span}`);
    if (component.layout.padding) layoutClasses.push(component.layout.padding);
    if (component.layout.margin) layoutClasses.push(component.layout.margin);
    if (component.layout.class_name)
      layoutClasses.push(component.layout.class_name);
  }

  return (
    <div className={cn(layoutClasses)} style={layoutStyle}>
      <Component {...component.props} id={component.id}>
        {component.children?.map((child, index) => (
          <DynamicRenderer key={child.id || index} component={child} />
        ))}
      </Component>
    </div>
  );
};

// --- Basic Components ---

const SDPage: React.FC<any> = ({ title, children }) => (
  <div className="space-y-6 animate-in fade-in duration-500">
    {title && <h1 className="text-3xl font-bold tracking-tight">{title}</h1>}
    {children}
  </div>
);

const SDCard: React.FC<any> = ({ title, children }) => (
  <Card>
    {title && (
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
    )}
    <CardContent>{children}</CardContent>
  </Card>
);

const SDGrid: React.FC<any> = ({ columns = 1, gap = 4, children }) => (
  <div
    className={`grid gap-${gap}`}
    style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
  >
    {children}
  </div>
);

const SDText: React.FC<any> = ({ content, className }) => (
  <p className={cn("text-sm text-muted-foreground", className)}>{content}</p>
);

const SDButton: React.FC<any> = ({ label, action, variant = "default" }) => (
  <Button variant={variant} onClick={() => console.log("Action:", action)}>
    {label}
  </Button>
);

const SDStats: React.FC<any> = ({ label, value, change }) => (
  <Card>
    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
      <CardTitle className="text-sm font-medium">{label}</CardTitle>
    </CardHeader>
    <CardContent>
      <div className="text-2xl font-bold">{value}</div>
      {change && <p className="text-xs text-muted-foreground">{change}</p>}
    </CardContent>
  </Card>
);

// Default Registry
export const defaultRegistry: Record<string, React.ComponentType<any>> = {
  page: SDPage,
  card: SDCard,
  grid: SDGrid,
  text: SDText,
  button: SDButton,
  stats: SDStats,
  chart: ChartWidget,
  container: ({ children }) => <div>{children}</div>,
};

export const SDUIRoot: React.FC<{
  component: ComponentData;
  registry?: Record<string, React.ComponentType<any>>;
}> = ({ component, registry = defaultRegistry }) => (
  <ComponentRegistryContext.Provider value={registry}>
    <DynamicRenderer component={component} />
  </ComponentRegistryContext.Provider>
);
