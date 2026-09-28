import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle, Trash2, Ban } from "lucide-react";
import { useState } from "react";

interface ConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: "default" | "destructive" | "warning";
  onConfirm: () => void | Promise<void>;
  loading?: boolean;
}

export function ConfirmationDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  variant = "destructive",
  onConfirm,
  loading = false,
}: ConfirmationDialogProps) {
  const [isConfirming, setIsConfirming] = useState(false);

  const handleConfirm = async () => {
    setIsConfirming(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } finally {
      setIsConfirming(false);
    }
  };

  const Icon = variant === "warning" ? AlertTriangle : variant === "destructive" ? Trash2 : Ban;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Icon className="h-5 w-5 text-destructive" />
            {title}
          </DialogTitle>
          <DialogDescription>
            {description}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-4">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={loading || isConfirming}
          >
            {cancelLabel}
          </Button>
          <Button
            variant={variant === "destructive" ? "destructive" : "default"}
            onClick={handleConfirm}
            disabled={loading || isConfirming}
          >
            {loading || isConfirming ? (
              <>
                <span className="animate-spin mr-2">⏳</span>
                Processing...
              </>
            ) : (
              confirmLabel
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function useConfirmation() {
  const [dialogState, setDialogState] = useState<{
    open: boolean;
    config: Omit<ConfirmationDialogProps, "open" | "onOpenChange"> | null;
  }>({
    open: false,
    config: null,
  });

  const confirm = (config: Omit<ConfirmationDialogProps, "open" | "onOpenChange">) => {
    setDialogState({ open: true, config });
  };

  const dialog = dialogState.config ? (
    <ConfirmationDialog
      open={dialogState.open}
      onOpenChange={(open) => setDialogState((prev) => ({ ...prev, open }))}
      {...dialogState.config}
    />
  ) : null;

  return { confirm, dialog };
}
