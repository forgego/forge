import type { useNavigate } from "@tanstack/react-router";
import type { FieldMetadata, RelationMetadata, Metadata } from "../../api/types";
import { TableCell } from "../ui/table";
import { EmptyValue } from "../ui/empty-state";
import { StatusBadge } from "../ui/status-badge";
import { CheckCircle, XCircle } from "lucide-react";
import { cn } from "../../lib/utils";

export type ListCellProps = {
  fieldName: string;
  colIdx: number;
  obj: Record<string, any>;
  field: FieldMetadata | undefined;
  relation: RelationMetadata | undefined;
  navigate: ReturnType<typeof useNavigate>;
  isSelected?: boolean;
  metadata?: Metadata;
  modelName?: string;
  display?: Record<string, Record<string, string>>;
};

export function ListCell({
  fieldName,
  colIdx,
  obj,
  field,
  relation,
  navigate,
  isSelected = false,
  metadata,
  modelName,
  display,
}: ListCellProps) {
  const val = obj[fieldName];
  const isPrimary = colIdx === 0;
  const isDate =
    field?.type === "date" ||
    field?.type === "datetime" ||
    fieldName.endsWith("_at") ||
    fieldName.endsWith("_date");
  const matchedChoice = field?.choices?.find(
    (c) => String(c.value) === String(val)
  );
  const labels = relation
    ? (display?.[relation.name] ?? display?.[fieldName])
    : undefined;
  const label = labels?.[String(val)];
  const hasLabel = typeof label === "string" && label.trim() !== "";
  const isEmpty =
    val === null || val === undefined || val === "";

  return (
    <TableCell
      className={cn(
        "py-3 text-ui font-medium text-foreground/80",
        isPrimary &&
          "sticky left-[var(--sticky-id-offset)] z-20",
        isPrimary &&
          (isSelected ? "bg-inherit" : "bg-surface-2")
      )}
    >
      {isEmpty ? (
        <EmptyValue />
      ) : fieldName === "active" ||
        typeof val === "boolean" ? (
        val ? (
          <StatusBadge tone="success">
            <CheckCircle className="h-3 w-3" aria-hidden /> Yes
          </StatusBadge>
        ) : (
          <StatusBadge tone="danger">
            <XCircle className="h-3 w-3" aria-hidden /> No
          </StatusBadge>
        )
      ) : matchedChoice ? (
        <StatusBadge tone="neutral">
          {matchedChoice.label}
        </StatusBadge>
      ) : isDate && val ? (
        <span className="font-mono text-meta tabular-nums text-muted-foreground">
          {(() => {
            try {
              const d = new Date(val);
              return isNaN(d.getTime())
                ? String(val)
                : d.toLocaleDateString(undefined, {
                    month: "short",
                    day: "numeric",
                    year: "numeric",
                  });
            } catch {
              return String(val);
            }
          })()}
        </span>
      ) : relation ? (
        hasLabel ? (
          <button
            type="button"
            data-testid={`fk-${fieldName}-${obj.id}`}
            onClick={() =>
              navigate({
                to: "/$model/$id/view",
                params: {
                  model: relation.related_model,
                  id: String(val),
                },
              })
            }
            title={label}
            className="text-ui truncate max-w-[16rem] transition-colors duration-fast ease-out hover:text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
          >
            {label}
            <span className="ml-1.5 font-mono text-meta text-muted-foreground tabular-nums">
              #{String(val)}
            </span>
          </button>
        ) : (
          <button
            type="button"
            data-testid={`fk-${fieldName}-${obj.id}`}
            onClick={() =>
              navigate({
                to: "/$model/$id/view",
                params: {
                  model: relation.related_model,
                  id: String(val),
                },
              })
            }
            title={`View related ${relation.label ?? relation.related_model}`}
            className="font-mono text-meta tabular-nums text-muted-foreground transition-colors duration-fast ease-out hover:text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
          >
            #{String(val)}
          </button>
        )
      ) : isPrimary ? (
        <button
          type="button"
          onClick={() =>
            navigate({
              to: metadata?.permissions?.view
                ? "/$model/$id/view"
                : "/$model/$id",
              params: { model: modelName ?? metadata?.name ?? "", id: obj.id },
            })
          }
          className="font-semibold text-foreground hover:text-primary hover:underline transition-colors text-left rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {val?.toString() || `#${obj.id}`}
        </button>
      ) : (
        String(val)
      )}
    </TableCell>
  );
}
