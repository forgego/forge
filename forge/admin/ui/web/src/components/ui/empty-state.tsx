import React from "react";
import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

export interface EmptyStateProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  icon?: LucideIcon
  title: React.ReactNode
  description?: React.ReactNode
  action?: React.ReactNode
}

export const EmptyState = React.forwardRef<HTMLDivElement, EmptyStateProps>(
  ({ icon: Icon, title, description, action, className, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn("flex flex-col items-center justify-center gap-3 px-canvas py-12 text-center", className)}
        {...props}
      >
        {Icon && (
          <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-border-subtle bg-surface-sunken text-muted-foreground">
            <Icon className="h-[18px] w-[18px]" aria-hidden />
          </div>
        )}
        <p className="text-lead text-foreground">{title}</p>
        {description && <p className="max-w-sm text-meta text-muted-foreground">{description}</p>}
        {action && <div className="pt-1">{action}</div>}
      </div>
    );
  }
);
EmptyState.displayName = "EmptyState";

export function EmptyValue({ className }: { className?: string }) {
  return (
    <span
      aria-label="No value"
      className={cn("select-none text-meta text-muted-foreground/60", className)}
    >
      —
    </span>
  );
}