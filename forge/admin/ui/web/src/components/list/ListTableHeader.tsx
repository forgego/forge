import type { FieldMetadata } from "../../api/types";
import { TableHeader, TableRow, TableHead } from "../ui/table";
import { Checkbox } from "../ui/checkbox";
import { ArrowUp, ArrowDown, ArrowUpDown } from "lucide-react";
import { cn } from "../../lib/utils";

export type ListTableHeaderProps = {
  allSelected: boolean;
  someSelected: boolean;
  onSelectAll: (checked: boolean) => void;
  displayFields: string[];
  fieldsByName: Map<string, FieldMetadata>;
  sortField: string | null;
  sortDir: "asc" | "desc" | null;
  onToggleSort: (fieldName: string) => void;
};

export function ListTableHeader({
  allSelected,
  someSelected,
  onSelectAll,
  displayFields,
  fieldsByName,
  sortField,
  sortDir,
  onToggleSort,
}: ListTableHeaderProps) {
  const sortIcon = (fieldName: string) => {
    if (sortField !== fieldName)
      return <ArrowUpDown className="h-3 w-3 opacity-50" aria-hidden />;
    if (sortDir === "asc") return <ArrowUp className="h-3 w-3" aria-hidden />;
    return <ArrowDown className="h-3 w-3" aria-hidden />;
  };

  return (
    <TableHeader className="bg-muted/40 sticky top-0">
      <TableRow className="hover:bg-transparent border-border-subtle">
        <TableHead className="w-[52px] pl-6 py-3 sticky left-0 z-20 bg-surface-sunken">
          <Checkbox
            data-testid="select-all"
            aria-label="Select all rows on this page"
            checked={allSelected}
            ref={(el) => {
              if (el) el.indeterminate = someSelected;
            }}
            onChange={(e) => onSelectAll(e.target.checked)}
          />
        </TableHead>
        {displayFields.map((fieldName, colIdx) => {
          const field = fieldsByName.get(fieldName);
          const active = sortField === fieldName;
          return (
            <TableHead
              key={fieldName}
              aria-sort={
                active
                  ? sortDir === "asc"
                    ? "ascending"
                    : "descending"
                  : "none"
              }
              className={cn(
                "py-3 min-w-[140px]",
                colIdx === 0 &&
                  "sticky left-[var(--sticky-id-offset)] z-20 bg-surface-sunken"
              )}
            >
              <button
                type="button"
                data-testid={`sort-${fieldName}`}
                onClick={() => onToggleSort(fieldName)}
                title={`Sort by ${field?.label || fieldName}`}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded px-1 py-0.5 -ml-1 text-meta font-semibold",
                  active
                    ? "text-foreground"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                {field?.label || fieldName}
                {sortIcon(fieldName)}
              </button>
            </TableHead>
          );
        })}
        <TableHead className="w-[110px] text-meta font-semibold text-muted-foreground text-right pr-6">
          Actions
        </TableHead>
      </TableRow>
    </TableHeader>
  );
}
