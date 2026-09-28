import type { ActionMetadata, PermissionMetadata } from "../../api/types";
import { Button } from "../ui/button";
import { Loader2, Zap, Trash2 } from "lucide-react";
import { cn } from "../../lib/utils";

export type ListBulkToolbarProps = {
  hasSelection: boolean;
  selectedIds: (string | number)[];
  setSelectedIds: (ids: (string | number)[]) => void;
  actions: ActionMetadata[];
  permissions: PermissionMetadata;
  handleActionClick: (action: ActionMetadata) => void;
  actionLoading: boolean;
  setBulkDeleteConfirmOpen: (open: boolean) => void;
  bulkDeletePending: boolean;
};

export function ListBulkToolbar({
  hasSelection,
  selectedIds,
  setSelectedIds,
  actions,
  permissions,
  handleActionClick,
  actionLoading,
  setBulkDeleteConfirmOpen,
  bulkDeletePending,
}: ListBulkToolbarProps) {
  return (
    <div
      data-testid="bulk-toolbar"
      className={cn(
        "flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 rounded-lg border px-4 py-3 text-body",
        hasSelection
          ? "bg-primary/[0.06] border-primary/20 text-foreground"
          : "bg-muted/40 border-border-subtle text-muted-foreground"
      )}
      aria-live="polite"
    >
      <div className="flex flex-wrap items-center gap-3">
        <span className="font-medium">
          {hasSelection
            ? <span className="font-mono tabular-nums">{selectedIds.length} selected</span>
            : "Select rows to enable bulk actions"}
        </span>
        <div className="h-4 w-px bg-border hidden sm:block" />
        <div className="flex flex-wrap items-center gap-2">
          {actions.map((action) => (
            <Button
              key={action.name}
              data-testid={`bulk-action-${action.name}`}
              variant="ghost"
              size="sm"
              className="h-8 px-3 text-meta font-medium"
              onClick={() => handleActionClick(action)}
              disabled={!hasSelection || actionLoading}
            >
              {actionLoading ? (
                <Loader2 className="h-3 w-3 animate-spin mr-2" />
              ) : (
                <Zap className="h-3 w-3 mr-2 text-muted-foreground" />
              )}
              {action.label}
            </Button>
          ))}
          {permissions.delete && (
            <Button
              data-testid="bulk-delete-button"
              variant="destructive"
              size="sm"
              className={cn(
                "h-8 px-3 text-meta font-medium gap-1.5",
                !hasSelection && "opacity-50"
              )}
              onClick={() => setBulkDeleteConfirmOpen(true)}
              disabled={!hasSelection || bulkDeletePending}
            >
              {bulkDeletePending ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <Trash2 className="h-3 w-3" />
              )}
              Delete (<span className="font-mono tabular-nums">{selectedIds.length}</span>)
            </Button>
          )}
        </div>
      </div>
      <Button
        data-testid="bulk-cancel"
        variant="ghost"
        size="sm"
        className="h-8 px-3 text-meta font-medium text-muted-foreground hover:text-foreground self-start sm:self-auto"
        onClick={() => setSelectedIds([])}
        disabled={!hasSelection}
      >
        Clear selection
      </Button>
    </div>
  );
}
