import * as React from "react";
import { Search, FileText, Loader2 } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { ModelIcon } from "../ModelIcon";

import type { ModelListMetadata, SearchResultGroup } from "../../api/types";
import { adminAPI } from "../../api/client";
import { useDebouncedValue } from "../../hooks/useDebouncedValue";
import { cn } from "../../lib/utils";

type GlobalSearchProps = {
  models?: ModelListMetadata[];
  compact?: boolean;
  triggerLabel?: string;
  className?: string;
};

type FlatItem =
  | { kind: "model"; key: string; label: string; sub?: string; icon?: string; model: ModelListMetadata }
  | { kind: "record"; key: string; label: string; sub?: string; url: string };

export function GlobalSearch({
  models = [],
  compact = false,
  triggerLabel = "Search...",
  className,
}: GlobalSearchProps) {
  const navigate = useNavigate();
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const [results, setResults] = React.useState<SearchResultGroup[]>([]);
  // Last query the server answered; loading is derived from it so no
  // synchronous setState-in-effect is needed.
  const [loadedQuery, setLoadedQuery] = React.useState<string | null>(null);
  const [activeIndex, setActiveIndex] = React.useState(0);
  const inputRef = React.useRef<HTMLInputElement>(null);
  const listRef = React.useRef<HTMLDivElement>(null);

  const debouncedQuery = useDebouncedValue(query.trim(), 250);
  const isLoading = open && debouncedQuery !== "" && loadedQuery !== debouncedQuery;

  const openPalette = React.useCallback(() => {
    setOpen(true);
    setActiveIndex(0);
    setLoadedQuery(null);
    requestAnimationFrame(() => inputRef.current?.focus());
  }, []);

  const closePalette = React.useCallback(() => {
    setOpen(false);
    setQuery("");
    setResults([]);
    setLoadedQuery(null);
    setActiveIndex(0);
  }, []);

  const onQueryChange = (value: string) => {
    setQuery(value);
    setActiveIndex(0);
    if (value.trim() === "") {
      setResults([]);
      setLoadedQuery(null);
    }
  };

  // Expose an imperative opener so other pages (dashboard Quick Find)
  // can launch the palette without synthesizing keyboard events.
  React.useEffect(() => {
    (window as any).__forgeOpenSearch = openPalette;
    return () => {
      delete (window as any).__forgeOpenSearch;
    };
  }, [openPalette]);

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && open) {
        e.preventDefault();
        closePalette();
        return;
      }
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        if (open) closePalette();
        else openPalette();
        return;
      }
      if (
        e.key === "/" &&
        !open &&
        !(e.target instanceof HTMLInputElement) &&
        !(e.target instanceof HTMLTextAreaElement) &&
        !(e.target as HTMLElement)?.isContentEditable
      ) {
        e.preventDefault();
        openPalette();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open, openPalette, closePalette]);

  React.useEffect(() => {
    if (!open || debouncedQuery === "") {
      return;
    }
    let cancelled = false;
    adminAPI
      .globalSearch({ query: debouncedQuery })
      .then((res) => {
        if (cancelled) return;
        setResults(res.results || []);
        setLoadedQuery(debouncedQuery);
        setActiveIndex(0);
      })
      .catch((err) => {
        if (cancelled) return;
        console.error("Search failed:", err);
        setLoadedQuery(debouncedQuery);
      });
    return () => {
      cancelled = true;
    };
  }, [open, debouncedQuery]);

  const matchingModels = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    if (q === "") return models;
    return models.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.verbose_name.toLowerCase().includes(q) ||
        m.verbose_name_plural.toLowerCase().includes(q)
    );
  }, [models, query]);

  const flatItems: FlatItem[] = React.useMemo(() => {
    const items: FlatItem[] = [];
    if (debouncedQuery !== "") {
      for (const group of results) {
        for (const item of group.items) {
          items.push({
            kind: "record",
            key: `record-${group.model}-${item.id}`,
            label: item.title,
            sub: group.model,
            url: item.url,
          });
        }
      }
    }
    for (const m of matchingModels.slice(0, 8)) {
      items.push({
        kind: "model",
        key: `model-${m.name}`,
        label: m.verbose_name_plural,
        sub: `${m.count} records`,
        icon: m.icon,
        model: m,
      });
    }
    return items;
  }, [results, matchingModels, debouncedQuery]);

  // Clamp instead of resetting in an effect: keeps selection valid when
  // the result list shrinks without cascading renders.
  const safeIndex = Math.min(activeIndex, Math.max(flatItems.length - 1, 0));
  const activeItem = flatItems[safeIndex];

  const activateItem = (item: FlatItem) => {
    if (item.kind === "model") {
      navigate({ to: "/$model", params: { model: item.model.name } });
    } else if (item.url) {
      const to = item.url.startsWith("/admin")
        ? item.url.replace("/admin", "") || "/"
        : item.url;
      navigate({ to } as any);
    }
    closePalette();
  };

  const onInputKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, flatItems.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      const item = activeItem;
      if (item) activateItem(item);
    } else if (e.key === "Escape") {
      e.preventDefault();
      closePalette();
    }
  };

  React.useEffect(() => {
    listRef.current
      ?.querySelector(`[data-index="${safeIndex}"]`)
      ?.scrollIntoView({ block: "nearest" });
  }, [safeIndex]);

  return (
    <div className={cn("relative", className)}>
      <button
        type="button"
        data-testid="global-search-trigger"
        className={cn(
          "flex items-center gap-2 rounded-md border border-border bg-background py-2 text-sm text-muted-foreground hover:text-foreground hover:border-foreground/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          compact
            ? "h-9 w-9 justify-center p-0"
            : "h-9 w-9 justify-center p-0 sm:w-full sm:justify-start sm:px-3"
        )}
        onClick={openPalette}
        aria-label="Open search"
        title="Search (Ctrl+K)"
      >
        <Search className="h-4 w-4 shrink-0" aria-hidden />
        {!compact && (
          <>
            <span className="hidden sm:block flex-1 text-left">{triggerLabel}</span>
            <kbd className="hidden md:inline-flex items-center rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono">
              ⌘K
            </kbd>
          </>
        )}
      </button>

      {open && (
        <div
          className="fixed inset-0 z-[100] bg-background/70 backdrop-blur-sm"
          onMouseDown={(e) => {
            if (e.target === e.currentTarget) closePalette();
          }}
        >
          <div className="flex items-start justify-center pt-[12vh] px-4">
            <div
              role="dialog"
              aria-modal="true"
              aria-label="Global search"
              className="w-full max-w-xl rounded-xl border border-border bg-card shadow-2xl overflow-hidden"
            >
              <div className="flex items-center gap-2 border-b border-border px-4 py-3">
                <Search className="h-4 w-4 text-muted-foreground shrink-0" aria-hidden />
                <input
                  ref={inputRef}
                  data-testid="global-search-input"
                  value={query}
                  onChange={(e) => onQueryChange(e.target.value)}
                  onKeyDown={onInputKeyDown}
                  placeholder="Search models and records..."
                  role="combobox"
                  aria-expanded="true"
                  aria-controls="global-search-results"
                  aria-activedescendant={
                    activeItem ? `gs-item-${safeIndex}` : undefined
                  }
                  autoComplete="off"
                  className="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
                />
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground shrink-0" aria-label="Searching" />
                ) : (
                  <kbd className="shrink-0 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground">
                    ESC
                  </kbd>
                )}
              </div>

              <div
                ref={listRef}
                id="global-search-results"
                role="listbox"
                aria-label="Search results"
                className="max-h-[55vh] overflow-y-auto p-2"
              >
                {query.trim() === "" && (
                  <div className="px-3 py-2">
                    <p className="pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                      Models
                    </p>
                    {matchingModels.length === 0 && (
                      <p className="px-1 py-3 text-xs text-muted-foreground">
                        No models registered.
                      </p>
                    )}
                  </div>
                )}

                {query.trim() !== "" &&
                  !isLoading &&
                  flatItems.length === 0 && (
                    <p className="p-4 text-xs text-muted-foreground">
                      No matching models or records for “{query.trim()}”.
                    </p>
                  )}

                {flatItems.map((item, index) => (
                  <button
                    key={item.key}
                    id={`gs-item-${index}`}
                    data-index={index}
                    data-testid={`global-search-result-${index}`}
                    type="button"
                    role="option"
                    aria-selected={index === safeIndex}
                    onClick={() => activateItem(item)}
                    onMouseEnter={() => setActiveIndex(index)}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-left",
                      index === safeIndex
                        ? "bg-accent text-accent-foreground"
                        : "text-foreground"
                    )}
                  >
                    {item.kind === "model" ? (
                      <ModelIcon name={item.icon} className="h-4 w-4 shrink-0 text-muted-foreground" />
                    ) : (
                      <FileText className="h-4 w-4 shrink-0 text-muted-foreground" aria-hidden />
                    )}
                    <span className="flex-1 truncate font-medium">
                      {item.label}
                    </span>
                    {item.sub && (
                      <span className="shrink-0 text-[11px] text-muted-foreground">
                        {item.sub}
                      </span>
                    )}
                  </button>
                ))}
              </div>

              <div className="flex items-center gap-4 border-t border-border px-4 py-2 text-[11px] text-muted-foreground">
                <span className="flex items-center gap-1">
                  <kbd className="rounded border border-border bg-muted px-1 font-mono">↑↓</kbd> navigate
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="rounded border border-border bg-muted px-1 font-mono">↵</kbd> open
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="rounded border border-border bg-muted px-1 font-mono">esc</kbd> close
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
