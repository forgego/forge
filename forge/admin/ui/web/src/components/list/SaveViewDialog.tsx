import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Loader2 } from "lucide-react";

export type SaveViewDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  name: string;
  onNameChange: (name: string) => void;
  onConfirm: () => void;
  pending?: boolean;
};

export function SaveViewDialog({
  open,
  onOpenChange,
  name,
  onNameChange,
  onConfirm,
  pending = false,
}: SaveViewDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" data-testid="save-view-dialog">
        <DialogHeader>
          <DialogTitle>Save current view</DialogTitle>
          <DialogDescription>
            Name this combination of search, filters and sorting to reuse it later.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <label htmlFor="save-view-name" className="text-ui font-medium">
            View name
          </label>
          <Input
            id="save-view-name"
            data-testid="save-view-name"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="e.g. Active products"
            maxLength={80}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                onConfirm();
              }
            }}
          />
        </div>
        <DialogFooter>
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            data-testid="save-view-confirm"
            onClick={onConfirm}
            disabled={!name.trim() || pending}
          >
            {pending ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : null}
            Save view
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
