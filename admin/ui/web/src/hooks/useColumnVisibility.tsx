import { useState } from 'react';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../components/ui/dropdown-menu';
import { Button } from '../components/ui/button';
import { Settings2 } from 'lucide-react';
import type { FieldMetadata } from '../api/types';

interface ColumnCustomizerProps {
  fields: FieldMetadata[];
  visibleColumns: string[];
  onColumnsChange: (columns: string[]) => void;
}

export function ColumnCustomizer({
  fields,
  visibleColumns,
  onColumnsChange,
}: ColumnCustomizerProps) {
  const allColumns = fields
    .filter((f) => !f.read_only)
    .map((f) => f.name);

  const toggleColumn = (column: string) => {
    if (visibleColumns.includes(column)) {
      // Don't remove the last visible column
      if (visibleColumns.length <= 1) return;
      onColumnsChange(visibleColumns.filter((c) => c !== column));
    } else {
      onColumnsChange([...visibleColumns, column]);
    }
  };

  const showAll = () => onColumnsChange(allColumns);
  const hideAll = () => {
    // Keep at least one column visible
    if (allColumns.length > 0) {
      onColumnsChange([allColumns[0]]);
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="h-8">
          <Settings2 className="h-3.5 w-3.5 mr-1.5" />
          Columns
          <span className="ml-1.5 text-xs text-muted-foreground">
            {visibleColumns.length}/{allColumns.length}
          </span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>Toggle Columns</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {allColumns.map((column) => {
          const field = fields.find((f) => f.name === column);
          const label = field?.label || column;
          const isVisible = visibleColumns.includes(column);

          return (
            <DropdownMenuCheckboxItem
              key={column}
              checked={isVisible}
              onCheckedChange={() => toggleColumn(column)}
              className="cursor-pointer"
            >
              {label}
            </DropdownMenuCheckboxItem>
          );
        })}
        <DropdownMenuSeparator />
        <DropdownMenuCheckboxItem
          checked={visibleColumns.length === allColumns.length}
          onCheckedChange={() => (visibleColumns.length === allColumns.length ? hideAll() : showAll())}
          className="cursor-pointer"
        >
          {visibleColumns.length === allColumns.length ? 'Hide All' : 'Show All'}
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

interface ColumnToggle {
  field: string;
  visible: boolean;
}

export function useColumnVisibility(fields: FieldMetadata[], defaultColumns?: string[]) {
  const [columns, setColumns] = useState<ColumnToggle[]>(() => {
    const defaultVisible = defaultColumns || fields.slice(0, 5).map((f) => f.name);
    return fields.map((field) => ({
      field: field.name,
      visible: defaultVisible.includes(field.name),
    }));
  });

  const visibleColumns = columns.filter((c) => c.visible).map((c) => c.field);

  const toggleColumn = (field: string) => {
    setColumns((prev) =>
      prev.map((c) => {
        if (c.field === field) {
          // Don't allow hiding the last visible column
          if (!c.visible && prev.filter((p) => p.visible).length >= prev.length - 1) {
            return c;
          }
          return { ...c, visible: !c.visible };
        }
        return c;
      })
    );
  };

  const setColumnVisibility = (field: string, visible: boolean) => {
    setColumns((prev) =>
      prev.map((c) => {
        if (c.field === field) {
          // Don't allow hiding the last visible column
          if (!visible && prev.filter((p) => p.visible).length === 1 && prev.find((p) => p.field === field)?.visible) {
            return c;
          }
          return { ...c, visible };
        }
        return c;
      })
    );
  };

  const resetToDefault = (defaultCols?: string[]) => {
    const defaultVisible = defaultCols || fields.slice(0, 5).map((f) => f.name);
    setColumns(fields.map((field) => ({
      field: field.name,
      visible: defaultVisible.includes(field.name),
    })));
  };

  return {
    columns,
    visibleColumns,
    toggleColumn,
    setColumnVisibility,
    resetToDefault,
  };
}
