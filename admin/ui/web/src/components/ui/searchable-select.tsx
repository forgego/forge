import * as React from "react";
import { Check, ChevronsUpDown, Search, Loader2 } from "lucide-react";
import { cn } from "../../lib/utils";
import { adminAPI } from "../../api/client";

interface SearchableSelectProps {
  model: string;
  value: any;
  onChange: (value: any) => void;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
}

export function SearchableSelect({
  model,
  value,
  onChange,
  placeholder = "Select item...",
  disabled = false,
}: SearchableSelectProps) {
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");
  const [options, setOptions] = React.useState<any[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const [selectedLabel, setSelectedLabel] = React.useState("");

  const containerRef = React.useRef<HTMLDivElement>(null);

  // Fetch initial label if value exists
  React.useEffect(() => {
    if (value && !selectedLabel) {
      const val = typeof value === 'object' ? value.id : value;
      adminAPI.getObject(model, val).then((obj: any) => {
        // Fallback label discovery
        setSelectedLabel(obj.name || obj.title || obj.label || obj.email || `ID: ${val}`);
      }).catch(() => {
        setSelectedLabel(`ID: ${value}`);
      });
    }
  }, [value, model]);

  // Search effect
  React.useEffect(() => {
    if (!open || disabled) return;

    const timer = setTimeout(() => {
      setIsLoading(true);
      adminAPI.autocomplete(model, "", search, 10)
      .then((res: any) => {
        setOptions(res.results || []);
      })
      .finally(() => {
        setIsLoading(false);
      });
    }, 300);

    return () => clearTimeout(timer);
  }, [search, model, open]);

  // Click outside handler
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
    if (
      containerRef.current &&
      !containerRef.current.contains(event.target as Node)
    ) {
      setOpen(false);
    }
  };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const currentValue = typeof value === 'object' ? value.id : value;

  return (
    <div className="relative w-full" ref={containerRef}>
      <button
        type="button"
        onClick={() => {
          if (disabled) return;
          setOpen(!open);
        }}
        className={cn(
          "flex h-10 w-full items-center justify-between rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all",
          !value && "text-muted-foreground"
        )}
        disabled={disabled}
      >
        <span className="truncate">{selectedLabel || placeholder}</span>
        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </button>

      {open && !disabled && (
        <div className="absolute z-50 mt-1 max-h-60 w-full overflow-hidden rounded-lg border border-border/50 bg-popover text-popover-foreground shadow-xl animate-in fade-in zoom-in-95 duration-100">
          <div className="flex items-center border-b border-border/50 px-3">
            <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
            <input
              autoFocus
              className="flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
              placeholder="Search..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              disabled={disabled}
            />
          </div>
          <div className="overflow-y-auto max-h-[200px] p-1">
            {isLoading && (
              <div className="flex items-center justify-center p-4">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}
            {!isLoading && options.length === 0 && (
              <div className="py-6 text-center text-sm text-muted-foreground">
                No items found.
              </div>
            )}
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                className={cn(
                  "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none hover:bg-accent hover:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
                  currentValue === option.value && "bg-accent/50"
                )}
                onClick={() => {
                  onChange(option.value);
                  setSelectedLabel(option.label);
                  setOpen(false);
                }}
              >
                <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                  {currentValue === option.value && (
                    <Check className="h-4 w-4" />
                  )}
                </span>
                <span className="truncate">{option.label}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
