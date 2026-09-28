import * as React from "react"
import { cn } from "@/lib/utils"

export interface PageHeaderProps extends Omit<React.HTMLAttributes<HTMLElement>, 'title'> {
  /** Small uppercase label above the title, e.g. the model group or section. */
  eyebrow?: React.ReactNode
  title: React.ReactNode
  description?: React.ReactNode
  /** Buttons rendered at the top right. */
  actions?: React.ReactNode
  /** Secondary facts rendered under the header, e.g. counts, last-updated. */
  meta?: React.ReactNode
}

const PageHeader = React.forwardRef<
  HTMLElement,
  PageHeaderProps
>(({ className, eyebrow, title, description, actions, meta, ...props }, ref) => (
  <header
    ref={ref}
    className={cn(
      "flex flex-col gap-3 border-b border-border-subtle pb-canvas",
      className
    )}
    {...props}
  >
    <div className="flex items-start justify-between gap-canvas">
      <div className="flex min-w-0 flex-col gap-1.5">
        {eyebrow && (
          <p className="text-micro text-muted-foreground">
            {eyebrow}
          </p>
        )}
        <h1 className="truncate text-title text-foreground">{title}</h1>
        {description && (
          <p className="max-w-2xl text-ui text-muted-foreground">
            {description}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex shrink-0 items-center gap-2">
          {actions}
        </div>
      )}
    </div>
    {meta && (
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-meta text-muted-foreground">
        {meta}
      </div>
    )}
  </header>
))

PageHeader.displayName = "PageHeader"

export { PageHeader }