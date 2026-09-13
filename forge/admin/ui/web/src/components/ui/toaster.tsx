import { Toaster as SonnerToaster } from "sonner"

export function Toaster() {
  return (
    <SonnerToaster
      position="bottom-right"
      closeButton
      toastOptions={{
        classNames: {
          toast:
            "rounded border border-border bg-surface-3 text-foreground shadow-overlay text-ui",
          description: "text-meta text-muted-foreground",
          actionButton: "rounded-sm bg-primary text-primary-foreground text-meta",
          cancelButton: "rounded-sm bg-surface-sunken text-foreground text-meta",
          error: "border-danger/30",
          success: "border-success/30",
        },
      }}
    />
  )
}
