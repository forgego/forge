import * as React from "react";
import { Search } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";

import type { ModelListMetadata } from "../../api/types";
import { adminAPI } from "../../api/client";
import { cn } from "../../lib/utils";

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
  const [results, setResults] = React.useState<any[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const [prevSearch, setPrevSearch] = React.useState({ open, query });
  if (prevSearch.open !== open || prevSearch.query !== query) {
    setPrevSearch({ open, query });
    if (!open || query.trim() === "") {
      setResults([]);
    }
  }
  const inputRef = React.useRef<HTMLInputElement>(null);
  const containerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(true);
        setTimeout(() => inputRef.current?.focus(), 0);
      }
      if (
        e.key === "/" &&
        !(e.target instanceof HTMLInputElement) &&
        !(e.target instanceof HTMLTextAreaElement) &&
        !(e.target as HTMLElement)?.isContentEditable
      ) {
        e.preventDefault();
        setOpen(true);
        setTimeout(() => inputRef.current?.focus(), 0);
      }
      if (e.key === "Escape") {
        setOpen(false);
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  React.useEffect(() => {
    if (!open || query.trim() === "") {
      return;
    }

    const timer = setTimeout(async () => {
      setIsLoading(true);
      try {
        const res = await adminAPI.globalSearch({ query });
        setResults(res.results || []);
      } catch (err) {
        console.error("Search failed:", err);
      } finally {
        setIsLoading(false);
      }
    }, 250);

    return () => clearTimeout(timer);
  }, [open, query]);

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

  const handleModelSelect = (model: ModelListMetadata) => {
    navigate({ to: "/$model", params: { model: model.name } });
    setOpen(false);
    setQuery("");
  };

  const handleResultSelect = (item: any) => {
    if (item.url) {
      navigate({ to: item.url.replace("/admin", "") });
    }
    setOpen(false);
    setQuery("");
  };

  return (
    <div className={cn("relative", className)}>
      <button
        type="button"
        className={cn(
          "flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm text-muted-foreground hover:text-foreground",
          compact ? "h-9 w-9 justify-center p-0" : "w-full"
        )}
        onClick={() => {
          setOpen(true);
          setTimeout(() => inputRef.current?.focus(), 0);
        }}
        aria-label="Open search"
      >
        <Search className="h-4 w-4" />
        {!compact && <span>{triggerLabel}</span>}
      </button>

      {open && (
        <div className="fixed inset-0 z-[100] bg-background/70 backdrop-blur-sm">
          <div className="flex items-start justify-center pt-[15vh] px-4">
            <div
              ref={containerRef}
              className="w-full max-w-xl rounded-xl border border-border bg-card shadow-xl"
            >
              <div className="flex items-center gap-2 border-b border-border px-4 py-3">
                <Search className="h-4 w-4 text-muted-foreground" />
                <input
                  ref={inputRef}
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search models and records..."
                  className="w-full bg-transparent text-sm outline-none"
                />
                {isLoading && (
                  <span className="text-xs text-muted-foreground">Loading</span>
                )}
              </div>

              <div className="max-h-[60vh] overflow-y-auto p-2">
                {query.trim() === "" && (
                  <div className="p-4 text-xs text-muted-foreground">
                    Start typing to search.
                  </div>
                )}

                {query.trim() !== "" && results.length === 0 && (
                  <div className="p-4 text-xs text-muted-foreground">
                    No matching records.
                  </div>
                )}

                {results.map((item, index) => (
                  <button
                    key={`${item.url}-${index}`}
                    type="button"
                    onClick={() => handleResultSelect(item)}
                    className="block w-full text-left rounded-md px-3 py-2 text-sm hover:bg-muted"
                  >
                    {item.label || item.title || item.url}
                  </button>
                ))}

                {query.trim() === "" && models.length > 0 && (
                  <div className="border-t border-border pt-2 mt-2">
                    <p className="px-3 pb-1 text-[11px] uppercase tracking-wide text-muted-foreground">
                      Models
                    </p>
                    {models.map((model) => (
                      <button
                        key={model.name}
                        type="button"
                        onClick={() => handleModelSelect(model)}
                        className="block w-full text-left rounded-md px-3 py-2 text-sm hover:bg-muted"
                      >
                        {model.verbose_name_plural}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
