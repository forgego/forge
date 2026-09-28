import * as React from "react";
import { Search, FileText, Loader2 } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { Command } from "cmdk";
import { ModelIcon } from "../ModelIcon";

import type { ModelListMetadata, SearchResultGroup } from "../../api/types";
import { adminAPI } from "../../api/client";
import { useDebouncedValue } from "../../hooks/useDebouncedValue";
import { cn } from "../../lib/utils";
import { stripAdminPrefix } from "../../lib/admin-prefix";

type GlobalSearchProps = {
  models?: ModelListMetadata[];
  compact?: boolean;
  triggerLabel?: string;
  className?: string;
};

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
  const inputRef = React.useRef<HTMLInputElement>(null);

  const debouncedQuery = useDebouncedValue(query.trim(), 250);
  const isLoading = open && debouncedQuery !== "" && loadedQuery !== debouncedQuery;

  const openPalette = React.useCallback(() => {
    setOpen(true);
    setLoadedQuery(null);
    requestAnimationFrame(() => inputRef.current?.focus());
  }, []);

  const closePalette = React.useCallback(() => {
    setOpen(false);
    setQuery("");
    setResults([]);
    setLoadedQuery(null);
  }, []);

  const onQueryChange = (value: string) => {
    setQuery(value);
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

  const modelLabelByName = React.useMemo(() => {
    const map = new Map<string, string>();
    (models ?? []).forEach((m: any) => {
      if (m?.name) map.set(m.name, m.verbose_name_plural ?? m.verbose_name ?? m.name);
    });
    return map;
  }, [models]);

  const recordItems = React.useMemo(() => {
    if (debouncedQuery === "") return [];
    const items: Array<{ key: string; label: string; sub: string; url: string }> = [];
    for (const group of results) {
      for (const item of group.items) {
        items.push({
          key: `record-${group.model}-${item.id}`,
          label: item.title,
          sub: modelLabelByName.get(group.model) ?? group.model,
          url: item.url,
        });
      }
    }
    return items;
  }, [results, debouncedQuery, modelLabelByName]);

  const modelItems = React.useMemo(() => matchingModels.slice(0, 8), [matchingModels]);

  const handleSelectModel = React.useCallback(
    (model: ModelListMetadata) => {
      navigate({ to: "/$model", params: { model: model.name } });
      closePalette();
    },
    [navigate, closePalette]
  );

  const handleSelectRecord = React.useCallback(
    (url: string) => {
      const to = stripAdminPrefix(url) || "/";
      navigate({ to } as any);
      closePalette();
    },
    [navigate, closePalette]
  );

  const handleInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Escape") {
      e.preventDefault();
      closePalette();
    }
  };

  return (
    <div className={cn("relative", className)}>
      <button
        type="button"
        data-testid="global-search-trigger"
        className={cn(
          "flex items-center gap-2 rounded border border-border bg-background py-2 text-ui text-muted-foreground hover:text-foreground hover:border-foreground/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
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
            <kbd className="hidden md:inline-flex items-center rounded border border-border bg-muted px-1.5 py-0.5 text-micro font-mono">
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
              className="w-full max-w-xl rounded border border-border bg-surface-3 shadow-dialog overflow-hidden"
            >
              <Command
                shouldFilter={false}
                className="[&_[cmdk-group-heading]]:px-3 [&_[cmdk-group-heading]]:py-1 [&_[cmdk-group-heading]]:text-micro [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-wider [&_[cmdk-group-heading]]:text-muted-foreground"
              >
                <div className="flex items-center gap-2 border-b border-border px-4 py-3">
                  <Search className="h-4 w-4 text-muted-foreground shrink-0" aria-hidden />
                  <Command.Input
                    ref={inputRef}
                    data-testid="global-search-input"
                    value={query}
                    onValueChange={onQueryChange}
                    onKeyDown={handleInputKeyDown}
                    placeholder="Search models and records..."
                    autoComplete="off"
                    className="w-full bg-transparent text-ui outline-none placeholder:text-muted-foreground"
                  />
                  {isLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin text-muted-foreground shrink-0" aria-label="Searching" />
                  ) : (
                    <kbd className="shrink-0 rounded border border-border bg-muted px-1.5 py-0.5 text-micro font-mono text-muted-foreground">
                      ESC
                    </kbd>
                  )}
                </div>

                <Command.List
                  id="global-search-results"
                  className="max-h-[55vh] overflow-y-auto p-2"
                >
                  {!isLoading && modelItems.length === 0 && recordItems.length === 0 && (
                    <Command.Empty className="p-4 text-meta text-muted-foreground">
                      {query.trim() === ""
                        ? "No models registered."
                        : `No matching models or records for “${query.trim()}”.`}
                    </Command.Empty>
                  )}
                  {isLoading && modelItems.length === 0 && recordItems.length === 0 && (
                    <Command.Empty className="p-4 text-meta text-muted-foreground">
                      {""}
                    </Command.Empty>
                  )}

                  {modelItems.length > 0 && (
                    <Command.Group heading="Models">
                      {modelItems.map((m) => (
                        <Command.Item
                          key={`model-${m.name}`}
                          value={`model-${m.name}`}
                          onSelect={() => handleSelectModel(m)}
                          className="flex w-full items-center gap-3 rounded px-3 py-2 text-ui text-left transition-colors duration-fast ease-out text-foreground data-[selected=true]:bg-surface-sunken data-[selected=true]:text-foreground aria-selected:bg-surface-sunken aria-selected:text-foreground"
                        >
                          <ModelIcon name={m.icon} className="h-4 w-4 shrink-0 text-muted-foreground" />
                          <span className="flex-1 truncate font-medium">{m.verbose_name_plural}</span>
                          <span className="shrink-0 font-mono text-meta tabular-nums text-muted-foreground">
                            {m.count} records
                          </span>
                        </Command.Item>
                      ))}
                    </Command.Group>
                  )}

                  {recordItems.length > 0 && (
                    <Command.Group heading="Records">
                      {recordItems.map((item) => (
                        <Command.Item
                          key={item.key}
                          value={item.key}
                          onSelect={() => handleSelectRecord(item.url)}
                          className="flex w-full items-center gap-3 rounded px-3 py-2 text-ui text-left transition-colors duration-fast ease-out text-foreground data-[selected=true]:bg-surface-sunken data-[selected=true]:text-foreground aria-selected:bg-surface-sunken aria-selected:text-foreground"
                        >
                          <FileText className="h-4 w-4 shrink-0 text-muted-foreground" aria-hidden />
                          <span className="flex-1 truncate font-medium">{item.label}</span>
                          <span className="shrink-0 font-mono text-meta tabular-nums text-muted-foreground">
                            {item.sub}
                          </span>
                        </Command.Item>
                      ))}
                    </Command.Group>
                  )}
                </Command.List>

                <div className="flex items-center gap-4 border-t border-border px-4 py-2 text-meta text-muted-foreground">
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
              </Command>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
