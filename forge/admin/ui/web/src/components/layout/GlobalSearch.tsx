import * as React from "react";
import { Search, Loader2, FileText, ArrowRight } from "lucide-react";
import { adminAPI } from "../../api/client";
import { useNavigate } from "@tanstack/react-router";

type GlobalSearchProps = {
  compact?: boolean;
  triggerLabel?: string;
  className?: string;
};

export function GlobalSearch({
  compact = false,
  triggerLabel = "Search...",
  className,
}: GlobalSearchProps) {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const [results, setResults] = React.useState<any[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const navigate = useNavigate();
  const containerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
      if (e.key === "Escape") {
        setOpen(false);
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  React.useEffect(() => {
    if (!open || !query) {
      setResults([]);
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
    }, 300);

    return () => clearTimeout(timer);
  }, [query, open]);

  // Click outside
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

  const handleSelect = (item: any) => {
    const path = item.url;
    navigate({ to: path });
    setOpen(false);
    setQuery("");
  };

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        aria-label="Open command palette"
        aria-keyshortcuts="Control+K Meta+K"
        className={
          compact
            ? `flex items-center justify-center h-9 w-9 rounded-md bg-muted/50 border border-border/50 text-muted-foreground hover:text-foreground hover:bg-muted transition-all ${className ?? ""}`
            : `flex items-center gap-2 px-3 py-1.5 rounded-full bg-muted/50 border border-border/50 text-muted-foreground hover:text-foreground hover:bg-muted transition-all group w-48 lg:w-64 ${className ?? ""}`
        }
      >
        <Search className="h-4 w-4" />
        {!compact && (
          <>
            <span className="text-sm font-medium">{triggerLabel}</span>
            <kbd className="ml-auto pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-background px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
              <span className="text-xs">⌘</span>K
            </kbd>
          </>
        )}
      </button>

      {open && (
        <div className="fixed inset-0 z-[100] bg-background/80 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="flex items-start justify-center pt-[15vh] px-4">
            <div
              ref={containerRef}
              className="w-full max-w-xl bg-card border border-border/50 rounded-2xl shadow-2xl shadow-black/20 overflow-hidden animate-in zoom-in-95 duration-200"
            >
              <div className="flex items-center p-4 border-b border-border/50">
                <Search className="h-5 w-5 text-muted-foreground mr-3" />
                <input
                  autoFocus
                  placeholder="Type to search across models..."
                  className="flex-1 bg-transparent border-none outline-none text-lg font-medium placeholder:text-muted-foreground"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
                {isLoading && (
                  <Loader2 className="h-5 w-5 animate-spin text-primary ml-2" />
                )}
              </div>

              <div className="max-h-[60vh] overflow-y-auto">
                {!query && (
                  <div className="p-8 text-center">
                    <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-3">
                      <FileText className="h-6 w-6 text-primary" />
                    </div>
                    <p className="text-sm font-bold text-foreground/80">
                      Search Forge Models
                    </p>
                    <p className="text-xs text-muted-foreground mt-1">
                      Start typing to see instantly correlated records
                    </p>
                  </div>
                )}

                {query && results.length === 0 && !isLoading && (
                  <div className="p-8 text-center text-muted-foreground">
                    No results found for "{query}"
                  </div>
                )}

                {results.map((group) => (
                  <div key={group.model} className="p-2">
                    <div className="px-3 py-2 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground/60">
                      {group.model}
                    </div>
                    {group.items.map((item: any) => (
                      <button
                        key={item.url}
                        onClick={() => handleSelect(item)}
                        className="w-full flex items-center justify-between p-3 rounded-xl hover:bg-accent hover:text-accent-foreground transition-all text-left group/item"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-lg bg-background border border-border/50 flex items-center justify-center">
                            <FileText className="h-4 w-4 text-muted-foreground" />
                          </div>
                          <span className="text-sm font-medium">
                            {item.title}
                          </span>
                        </div>
                        <ArrowRight className="h-4 w-4 opacity-0 group-hover/item:opacity-100 -translate-x-2 group-hover/item:translate-x-0 transition-all text-primary" />
                      </button>
                    ))}
                  </div>
                ))}
              </div>

              <div className="p-3 bg-muted/20 border-t border-border/10 flex items-center justify-between">
                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                  Forge Deep Search Engine
                </p>
                <div className="flex gap-2">
                  <span className="text-[10px] bg-background px-1.5 py-0.5 rounded border border-border/50 text-muted-foreground">
                    ESC to close
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
