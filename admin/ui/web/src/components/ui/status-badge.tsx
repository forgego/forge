import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const statusBadgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-sm border px-1.5 py-0.5 text-micro whitespace-nowrap",
  {
    variants: {
      tone: {
        neutral: "border-neutral/20 bg-neutral-surface text-neutral",
        success: "border-success/20 bg-success-surface text-success",
        warning: "border-warning/20 bg-warning-surface text-warning",
        danger: "border-danger/20 bg-danger-surface text-danger",
        info: "border-info/20 bg-info-surface text-info",
        accent: "border-primary/20 bg-primary/10 text-primary",
      },
    },
    defaultVariants: {
      tone: "neutral",
    },
  }
)

export interface StatusBadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof statusBadgeVariants> {
  /** Render a small leading dot in the current text color. */
  dot?: boolean
}

const StatusBadge = React.forwardRef<HTMLSpanElement, StatusBadgeProps>(
  ({ className, tone, dot, ...props }, ref) => {
    return (
      <span
        ref={ref}
        className={cn(statusBadgeVariants({ tone, className }))}
        {...props}
      >
        {dot && (
          <span aria-hidden="true" className="h-1.5 w-1.5 rounded-full bg-current" />
        )}
        {props.children}
      </span>
    )
  }
)

StatusBadge.displayName = "StatusBadge"

export { StatusBadge, statusBadgeVariants }