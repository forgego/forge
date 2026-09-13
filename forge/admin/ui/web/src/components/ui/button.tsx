import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
    "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded text-ui font-medium transition-colors duration-fast ease-out focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
    {
      variants: {
        variant: {
          default: "bg-primary text-primary-foreground hover:bg-primary/90",
          destructive: "bg-danger text-danger-fg hover:bg-danger/90",
          outline: "border border-border bg-transparent hover:bg-surface-sunken hover:text-foreground",
          secondary: "bg-surface-sunken text-foreground hover:bg-surface-sunken/70",
          ghost: "hover:bg-surface-sunken hover:text-foreground",
          link: "text-primary underline-offset-4 hover:underline",
        },
        size: {
          default: "h-9 px-3 py-2",
          sm: "h-command rounded-sm px-2 text-meta",
          lg: "h-10 rounded px-6",
          icon: "h-9 w-9",
        },
      },
      defaultVariants: {
        variant: "default",
        size: "default",
      },
    }
  )

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
