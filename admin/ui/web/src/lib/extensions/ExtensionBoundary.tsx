import React, { Component, type ErrorInfo, type ReactNode } from "react";
import { AlertCircle } from "lucide-react";

export interface ExtensionBoundaryProps {
  children?: ReactNode;
  fallback?: ReactNode | ((error: Error) => ReactNode);
  extensionName?: string;
  componentName?: string;
}

interface ExtensionBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ExtensionBoundary extends Component<
  ExtensionBoundaryProps,
  ExtensionBoundaryState
> {
  constructor(props: ExtensionBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ExtensionBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error(
      `[ExtensionBoundary] Error in extension component${
        this.props.extensionName ? ` (${this.props.extensionName})` : ""
      }:`,
      error,
      errorInfo
    );
  }

  render() {
    if (this.state.hasError) {
      if (typeof this.props.fallback === "function") {
        return this.props.fallback(this.state.error || new Error("Unknown error"));
      }
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <div
          role="alert"
          data-testid="extension-error-fallback"
          className="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive"
        >
          <div className="flex items-center gap-2 font-semibold">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>
              Extension Component Error
              {this.props.extensionName ? ` (${this.props.extensionName})` : ""}
            </span>
          </div>
          <p className="mt-1 text-xs opacity-90">
            {this.state.error?.message ||
              "An error occurred while rendering this extension component."}
          </p>
        </div>
      );
    }

    return this.props.children;
  }
}

export function withExtensionBoundary<P extends object>(
  Component: React.ComponentType<P>,
  options?: {
    extensionName?: string;
    componentName?: string;
    fallback?: ReactNode | ((error: Error) => ReactNode);
  }
): React.FC<P> {
  const Wrapped: React.FC<P> = (props: P) => (
    <ExtensionBoundary {...options}>
      <Component {...props} />
    </ExtensionBoundary>
  );
  Wrapped.displayName = `ExtensionBoundary(${
    Component.displayName || Component.name || "Component"
  })`;
  return Wrapped;
}
