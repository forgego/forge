import React from "react";
import { widgetRegistry } from "../widgets";
import { withExtensionBoundary } from "./ExtensionBoundary";

// UI design-system primitives exposed to extensions
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Badge } from "../../components/ui/badge";
import { StatusBadge } from "../../components/ui/status-badge";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "../../components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from "../../components/ui/table";
import {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from "../../components/ui/tabs";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "../../components/ui/dialog";
import {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
} from "../../components/ui/select";
import { Checkbox } from "../../components/ui/checkbox";
import { Switch } from "../../components/ui/switch";
import { Textarea } from "../../components/ui/textarea";
import { PageHeader } from "../../components/ui/page-header";
import { EmptyState } from "../../components/ui/empty-state";
import { Skeleton } from "../../components/ui/skeleton";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "../../components/ui/tooltip";

export const FORGE_ADMIN_API_VERSION = 1;

export interface PageNavConfig {
  label?: string;
  icon?: string | React.ComponentType<any>;
  order?: number;
  group?: string;
}

export interface RegisterPageOptions {
  path: string;
  title: string;
  nav?: PageNavConfig | boolean;
  extension?: string;
}

export interface RegisteredPage {
  name: string;
  path: string;
  fullPath: string;
  title: string;
  nav?: PageNavConfig;
  component: React.ComponentType<any>;
}

export interface FieldEditorTarget {
  model: string;
  field: string;
}

export interface ForgeAdminAPI {
  version: number;
  registerWidget: (type: string, Component: React.ComponentType<any>) => void;
  registerPage: (
    options: RegisterPageOptions,
    Component: React.ComponentType<any>
  ) => void;
  registerFieldEditor: (
    target: FieldEditorTarget,
    Component: React.ComponentType<any>
  ) => void;
  registerActionForm: (
    actionName: string,
    Component: React.ComponentType<any>
  ) => void;
  react: typeof React;
  ui: Record<string, any>;
}

declare global {
  interface Window {
    ForgeAdmin?: ForgeAdminAPI;
  }
}

// Internal registries
const pageRegistry = new Map<string, RegisteredPage>();
const fieldEditorRegistry = new Map<string, React.ComponentType<any>>();
const actionFormRegistry = new Map<string, React.ComponentType<any>>();

let activeExtensionName: string | null = null;

export function setCurrentExtensionContext(name: string | null): void {
  activeExtensionName = name;
}

export function getCurrentExtensionContext(): string | null {
  return activeExtensionName;
}

function normalizeExtPath(extensionName: string, rawPath: string): string {
  const trimmed = (rawPath || "").trim();
  if (trimmed.startsWith("/ext/")) {
    return trimmed.replace(/\/+$/, "") || trimmed;
  }
  const clean = trimmed.replace(/^\/+/, "").replace(/\/+$/, "");
  const cleanExt = (extensionName || "").trim().replace(/^\/+|\/+$/g, "");
  if (!cleanExt) {
    return clean ? `/ext/${clean}` : "/ext";
  }
  return clean ? `/ext/${cleanExt}/${clean}` : `/ext/${cleanExt}`;
}

export function registerPage(
  options: RegisterPageOptions,
  Component: React.ComponentType<any>
): void {
  const extensionName =
    options.extension || activeExtensionName || "custom";
  const fullPath = normalizeExtPath(extensionName, options.path);

  let navConfig: PageNavConfig | undefined = undefined;
  if (options.nav === true) {
    navConfig = { label: options.title };
  } else if (typeof options.nav === "object" && options.nav !== null) {
    navConfig = {
      label: options.nav.label || options.title,
      ...options.nav,
    };
  }

  const WrappedComponent = withExtensionBoundary(Component, {
    extensionName,
    componentName: `Page:${options.title || options.path}`,
  });

  const entry: RegisteredPage = {
    name: extensionName,
    path: options.path,
    fullPath,
    title: options.title,
    nav: navConfig,
    component: WrappedComponent,
  };

  pageRegistry.set(fullPath, entry);
}

export function getRegisteredPages(): RegisteredPage[] {
  return Array.from(pageRegistry.values());
}

export function getRegisteredPage(path: string): RegisteredPage | undefined {
  const normalized = (path || "").trim().replace(/\/+$/, "");
  const withExt = normalized.startsWith("/ext")
    ? normalized
    : `/ext/${normalized.replace(/^\/+/, "")}`;
  return (
    pageRegistry.get(withExt) ||
    pageRegistry.get(normalized) ||
    Array.from(pageRegistry.values()).find(
      (p) => p.fullPath === withExt || p.fullPath === normalized
    )
  );
}

export function clearRegisteredPages(): void {
  pageRegistry.clear();
}

export function registerFieldEditor(
  target: FieldEditorTarget,
  Component: React.ComponentType<any>
): void {
  const key = `${target.model}.${target.field}`;
  const WrappedComponent = withExtensionBoundary(Component, {
    extensionName: activeExtensionName || undefined,
    componentName: `FieldEditor:${key}`,
  });
  fieldEditorRegistry.set(key, WrappedComponent);
}

export function getFieldEditor(
  model: string,
  field: string
): React.ComponentType<any> | undefined {
  return fieldEditorRegistry.get(`${model}.${field}`);
}

export function getRegisteredFieldEditors(): Map<
  string,
  React.ComponentType<any>
> {
  return new Map(fieldEditorRegistry);
}

export function clearRegisteredFieldEditors(): void {
  fieldEditorRegistry.clear();
}

export function registerActionForm(
  actionName: string,
  Component: React.ComponentType<any>
): void {
  const WrappedComponent = withExtensionBoundary(Component, {
    extensionName: activeExtensionName || undefined,
    componentName: `ActionForm:${actionName}`,
  });
  actionFormRegistry.set(actionName, WrappedComponent);
}

export function getActionForm(
  actionName: string
): React.ComponentType<any> | undefined {
  return actionFormRegistry.get(actionName);
}

export function getRegisteredActionForms(): Map<
  string,
  React.ComponentType<any>
> {
  return new Map(actionFormRegistry);
}

export function clearRegisteredActionForms(): void {
  actionFormRegistry.clear();
}

export function registerWidget(
  type: string,
  Component: React.ComponentType<any>
): void {
  const WrappedComponent = withExtensionBoundary(Component, {
    extensionName: activeExtensionName || undefined,
    componentName: `Widget:${type}`,
  });
  widgetRegistry.register(type, WrappedComponent);
}

const uiPrimitives = {
  Button,
  Input,
  Label,
  Badge,
  StatusBadge,
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
  Checkbox,
  Switch,
  Textarea,
  PageHeader,
  EmptyState,
  Skeleton,
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
};

export const ForgeAdmin: ForgeAdminAPI = Object.freeze({
  version: FORGE_ADMIN_API_VERSION,
  registerWidget,
  registerPage,
  registerFieldEditor,
  registerActionForm,
  react: React,
  ui: Object.freeze(uiPrimitives),
});

// Mount on window object at module load
if (typeof window !== "undefined") {
  window.ForgeAdmin = ForgeAdmin;
}

export default ForgeAdmin;
