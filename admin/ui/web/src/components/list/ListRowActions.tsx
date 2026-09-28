import type { useNavigate } from "@tanstack/react-router";
import type { PermissionMetadata } from "../../api/types";
import { TableCell } from "../ui/table";
import { Button } from "../ui/button";
import { Eye, Edit, Trash2 } from "lucide-react";

export type ListRowActionsProps = {
  id: string | number;
  modelName: string;
  permissions: PermissionMetadata;
  navigate: ReturnType<typeof useNavigate>;
  onDelete: () => void;
  deletePending?: boolean;
};

export function ListRowActions({
  id,
  modelName,
  permissions,
  navigate,
  onDelete,
  deletePending = false,
}: ListRowActionsProps) {
  return (
    <TableCell className="text-right pr-4">
      <div className="flex justify-end gap-1 opacity-100 focus-within:opacity-100 lg:opacity-0 lg:group-hover:opacity-100 lg:group-focus-within:opacity-100 transition-opacity">
        {permissions.view && (
          <Button
            variant="ghost"
            size="icon"
            data-testid={`view-${id}`}
            className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10"
            onClick={() =>
              navigate({
                to: "/$model/$id/view",
                params: { model: modelName, id: String(id) },
              })
            }
            title="View record details"
            aria-label={`View record ${id}`}
          >
            <Eye className="h-3.5 w-3.5" />
          </Button>
        )}
        {permissions.change && (
          <Button
            variant="ghost"
            size="icon"
            data-testid={`edit-${id}`}
            className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10"
            onClick={() =>
              navigate({
                to: "/$model/$id",
                params: { model: modelName, id: String(id) },
              })
            }
            title="Edit record"
            aria-label={`Edit record ${id}`}
          >
            <Edit className="h-3.5 w-3.5" />
          </Button>
        )}
        {permissions.delete && (
          <Button
            variant="ghost"
            size="icon"
            data-testid={`delete-${id}`}
            className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
            onClick={onDelete}
            disabled={deletePending}
            title="Delete record"
            aria-label={`Delete record ${id}`}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        )}
      </div>
    </TableCell>
  );
}
