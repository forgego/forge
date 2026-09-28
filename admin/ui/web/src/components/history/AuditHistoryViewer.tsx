import { useState, useMemo } from "react";
import type { HistoryEntry } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  History,
  Clock,
  PlusCircle,
  RefreshCw,
  Trash2,
  ArrowRight,
  Search,
  SlidersHorizontal,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface AuditHistoryViewerProps {
  entries?: HistoryEntry[];
  isLoading?: boolean;
  modelName?: string;
  objectId?: string | number;
}

interface ParsedFieldChange {
  field: string;
  oldValue?: string;
  newValue?: string;
  rawText?: string;
}

function parseChangeStats(changeStats?: string): ParsedFieldChange[] {
  if (!changeStats) return [];

  try {
    const parsed = JSON.parse(changeStats);

    if (Array.isArray(parsed)) {
      return parsed.map((item) => ({
        field: "change",
        rawText: String(item),
      }));
    }

    if (parsed && typeof parsed === "object") {
      return Object.entries(parsed).map(([field, details]) => {
        if (details && typeof details === "object") {
          const detailObj = details as Record<string, any>;
          const oldVal =
            detailObj.from ?? detailObj.old ?? detailObj.before ?? undefined;
          const newVal =
            detailObj.to ?? detailObj.new ?? detailObj.after ?? undefined;

          return {
            field,
            oldValue: oldVal !== undefined ? String(oldVal) : undefined,
            newValue: newVal !== undefined ? String(newVal) : undefined,
          };
        }
        return {
          field,
          rawText: String(details),
        };
      });
    }

    return [{ field: "note", rawText: String(parsed) }];
  } catch {
    // If not JSON, check for "field: old -> new" patterns
    const arrowMatch = changeStats.split(";").map((part) => part.trim());
    return arrowMatch.map((item) => {
      const match = item.match(/^([^:]+):\s*(.*)\s*->\s*(.*)$/);
      if (match) {
        return {
          field: match[1].trim(),
          oldValue: match[2].trim(),
          newValue: match[3].trim(),
        };
      }
      return { field: "note", rawText: item };
    });
  }
}

function formatRelativeTime(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return timestamp;

    const now = new Date();
    const diffSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffSeconds < 60) return "just now";
    if (diffSeconds < 3600) {
      const mins = Math.floor(diffSeconds / 60);
      return `${mins}m ago`;
    }
    if (diffSeconds < 86400) {
      const hours = Math.floor(diffSeconds / 3600);
      return `${hours}h ago`;
    }
    if (diffSeconds < 604800) {
      const days = Math.floor(diffSeconds / 86400);
      return `${days}d ago`;
    }
    return date.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
    });
  } catch {
    return timestamp;
  }
}

function formatExactDate(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return timestamp;
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return timestamp;
  }
}

export function AuditHistoryViewer({
  entries = [],
  isLoading = false,
  modelName,
  objectId,
}: AuditHistoryViewerProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedActionFilter, setSelectedActionFilter] = useState<string>("all");

  const filteredEntries = useMemo(() => {
    return entries.filter((entry) => {
      // Action filter
      if (selectedActionFilter !== "all") {
        if ((entry.action || "").toLowerCase() !== selectedActionFilter.toLowerCase()) {
          return false;
        }
      }

      // Search query
      if (searchQuery.trim() !== "") {
        const query = searchQuery.toLowerCase();
        const user = (entry.user_name || entry.user_id || "system").toLowerCase();
        const action = (entry.action || "").toLowerCase();
        const stats = (entry.change_stats || "").toLowerCase();
        const repr = (entry.object_repr || "").toLowerCase();

        return (
          user.includes(query) ||
          action.includes(query) ||
          stats.includes(query) ||
          repr.includes(query)
        );
      }

      return true;
    });
  }, [entries, selectedActionFilter, searchQuery]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 space-y-4">
        <div className="relative flex items-center justify-center">
          <RefreshCw className="h-8 w-8 text-primary animate-spin" />
        </div>
        <p className="text-sm font-medium text-muted-foreground animate-pulse">
          Loading audit trail history...
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header and Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-2 border-b border-border/40">
        <div>
          <div className="flex items-center gap-2">
            <History className="h-5 w-5 text-primary" />
            <h3 className="text-lg font-semibold tracking-tight text-foreground">
              Audit Trail & Change History
            </h3>
            <Badge variant="secondary" className="text-xs font-mono">
              {entries.length} {entries.length === 1 ? "event" : "events"}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            Immutable log of creations, modifications, and administrative operations.
          </p>
        </div>

        {/* Filter controls */}
        <div className="flex flex-wrap items-center gap-2">
          <div className="relative w-full md:w-56">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search history..."
              className="h-8 pl-8 text-xs bg-background/50 border-border/60"
            />
          </div>

          <div className="flex items-center gap-1 bg-muted/60 p-0.5 rounded-lg border border-border/50">
            <Button
              variant={selectedActionFilter === "all" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-meta font-medium"
              onClick={() => setSelectedActionFilter("all")}
            >
              All
            </Button>
            <Button
              variant={selectedActionFilter === "add" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-meta font-medium text-success"
              onClick={() => setSelectedActionFilter("add")}
            >
              Created
            </Button>
            <Button
              variant={selectedActionFilter === "change" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-meta font-medium text-info"
              onClick={() => setSelectedActionFilter("change")}
            >
              Updated
            </Button>
            <Button
              variant={selectedActionFilter === "delete" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-meta font-medium text-danger"
              onClick={() => setSelectedActionFilter("delete")}
            >
              Deleted
            </Button>
          </div>
        </div>
      </div>

      {/* History Items */}
      {filteredEntries.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border/70 p-12 text-center bg-card/40">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted/60">
            <Clock className="h-6 w-6 text-muted-foreground" />
          </div>
          <h4 className="mt-4 text-sm font-semibold text-foreground">
            {entries.length === 0
              ? "No audit history recorded yet"
              : "No matching history entries"}
          </h4>
          <p className="mt-1 text-xs text-muted-foreground max-w-sm mx-auto">
            {entries.length === 0
              ? `All state transitions for ${modelName || "this record"} #${objectId || ""} will be captured automatically by the Forge Audit System.`
              : "Try adjusting your search query or action filter."}
          </p>
          {searchQuery && (
            <Button
              variant="outline"
              size="sm"
              className="mt-4 h-8 text-xs"
              onClick={() => {
                setSearchQuery("");
                setSelectedActionFilter("all");
              }}
            >
              Clear filters
            </Button>
          )}
        </div>
      ) : (
        <div className="relative pl-6 space-y-6 before:absolute before:left-3 before:top-2 before:bottom-2 before:w-[2px] before:bg-border/60">
          {filteredEntries.map((entry, index) => {
            const action = (entry.action || "change").toLowerCase();
            const isAdd = action === "add" || action === "create";
            const isDelete = action === "delete";
            const userLabel = entry.user_name || entry.user_id || "System Administrator";
            const changes = parseChangeStats(entry.change_stats);

            return (
              <div key={entry.id || index} className="relative group">
                {/* Timeline node icon */}
                <div
                  className={cn(
                    "absolute -left-[27px] top-1.5 flex h-6 w-6 items-center justify-center rounded-full border shadow-sm transition-transform group-hover:scale-110",
                    isAdd
                      ? "border-success/20 bg-success-surface text-success"
                      : isDelete
                      ? "border-danger/20 bg-danger-surface text-danger"
                      : "border-info/20 bg-info-surface text-info"
                  )}
                >
                  {isAdd ? (
                    <PlusCircle className="h-3.5 w-3.5" />
                  ) : isDelete ? (
                    <Trash2 className="h-3.5 w-3.5" />
                  ) : (
                    <RefreshCw className="h-3.5 w-3.5" />
                  )}
                </div>

                {/* Event Card */}
                <div className="rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm p-4 shadow-sm transition-all hover:border-border hover:shadow-md">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div className="flex items-center gap-2 flex-wrap">
                      <div className="flex items-center gap-1.5 font-medium text-xs text-foreground">
                        <div className="h-5 w-5 rounded-full bg-muted/80 flex items-center justify-center text-micro font-bold text-muted-foreground border border-border/40">
                          {userLabel.charAt(0).toUpperCase()}
                        </div>
                        <span>{userLabel}</span>
                      </div>

                      <Badge
                        variant="outline"
                        className={cn(
                          "text-micro font-semibold uppercase tracking-wider px-2 py-0.5",
                          isAdd && "border-success/20 bg-success-surface text-success",
                          isDelete && "border-danger/20 bg-danger-surface text-danger",
                          !isAdd && !isDelete && "border-info/20 bg-info-surface text-info"
                        )}
                      >
                        {isAdd ? "Created" : isDelete ? "Deleted" : "Updated"}
                      </Badge>

                      {entry.object_repr && (
                        <span className="text-xs text-muted-foreground truncate max-w-xs font-mono">
                          ({entry.object_repr})
                        </span>
                      )}
                    </div>

                    <div className="flex items-center gap-2 text-xs text-muted-foreground shrink-0">
                      <Clock className="h-3 w-3" />
                      <span title={formatExactDate(entry.timestamp)}>
                        {formatRelativeTime(entry.timestamp)}
                      </span>
                    </div>
                  </div>

                  {/* Changes Diff Display */}
                  {changes.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-border/40 space-y-2">
                      <div className="flex items-center gap-1.5 text-meta font-semibold text-muted-foreground uppercase tracking-wider">
                        <SlidersHorizontal className="h-3 w-3" />
                        <span>Modified Properties ({changes.length})</span>
                      </div>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                        {changes.map((change, cIdx) => (
                          <div
                            key={cIdx}
                            className="rounded-lg bg-muted/30 border border-border/40 px-3 py-2 text-xs flex flex-col gap-1"
                          >
                            <span className="font-mono font-semibold text-foreground/90">
                              {change.field}
                            </span>
                            {change.rawText ? (
                              <span className="text-muted-foreground break-all">
                                {change.rawText}
                              </span>
                            ) : (
                              <div className="flex items-center gap-2 text-xs">
                                <span className="line-through text-destructive/80 font-mono bg-destructive/10 px-1.5 py-0.5 rounded text-meta truncate max-w-[140px]">
                                  {change.oldValue ?? "null"}
                                </span>
                                <ArrowRight className="h-3 w-3 text-muted-foreground shrink-0" />
                                <span className="text-success font-mono bg-success-surface px-1.5 py-0.5 rounded text-meta font-medium truncate max-w-[140px]">
                                  {change.newValue ?? "null"}
                                </span>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
