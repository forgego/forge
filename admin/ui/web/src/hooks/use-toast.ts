import * as React from "react"
import { toast as sonnerToast } from "sonner"

export interface ToastOptions {
  title?: React.ReactNode
  description?: React.ReactNode
  variant?: "default" | "destructive" | "success"
  duration?: number
  action?: { label: string; onClick: () => void }
}

export function toast({ title, description, variant, duration, action }: ToastOptions) {
  const message = title ?? description ?? ""
  const opts = {
    description: title ? description : undefined,
    duration,
    action: action ? { label: action.label, onClick: action.onClick } : undefined,
  }
  if (variant === "destructive") return sonnerToast.error(message, opts)
  if (variant === "success") return sonnerToast.success(message, opts)
  return sonnerToast(message, opts)
}

export function useToast() {
  return { toast, dismiss: sonnerToast.dismiss }
}
