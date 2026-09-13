import { Card, CardContent } from "../ui/card";
import { Loader2 } from "lucide-react";

export const formatTimestamp = (timestamp?: string) => {
  if (!timestamp) return "Unknown time";
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return timestamp;
  return date.toLocaleString();
};

export const formatChangeStats = (changeStats?: string) => {
  if (!changeStats) return [];
  try {
    const parsed = JSON.parse(changeStats);
    if (Array.isArray(parsed)) {
      return parsed.map((item) => String(item));
    }
    if (parsed && typeof parsed === "object") {
      return Object.entries(parsed).map(([field, details]) => {
        if (details && typeof details === "object") {
          const detailObj = details as Record<string, any>;
          const fromValue =
            detailObj.from ?? detailObj.old ?? detailObj.before;
          const toValue = detailObj.to ?? detailObj.new ?? detailObj.after;
          if (fromValue !== undefined || toValue !== undefined) {
            return `${field}: ${fromValue ?? "∅"} → ${toValue ?? "∅"}`;
          }
        }
        return `${field}: ${String(details)}`;
      });
    }
    return [String(parsed)];
  } catch (error) {
    return [changeStats];
  }
};

export const getActionLabel = (action?: string) => {
  switch (action) {
    case "add":
      return "Created";
    case "change":
      return "Updated";
    case "delete":
      return "Deleted";
    default:
      return action ? action : "Updated";
  }
};

interface HistorySectionProps {
  historyData: any;
  historyLoading: boolean;
}

export function HistorySection({
  historyData,
  historyLoading,
}: HistorySectionProps) {
  return (
    <Card className="overflow-hidden max-w-4xl mx-auto">
      <CardContent className="p-4 sm:p-8 space-y-6">
        <div>
          <h2 className="text-xl font-semibold text-foreground/90">
            History
          </h2>
          <p className="text-sm text-muted-foreground">
            Track who changed this record and what was updated.
          </p>
        </div>
        {historyLoading ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Loading history…</span>
          </div>
        ) : historyData?.entries?.length ? (
          <ul className="relative border-l border-border/50 pl-6 space-y-6">
            {historyData.entries.map((entry: any) => {
              const changes = formatChangeStats(entry.change_stats);
              const userLabel =
                entry.user_name ||
                entry.user_id ||
                "System";
              return (
                <li key={entry.id} className="relative">
                  <span className="absolute -left-[9px] top-1.5 h-4 w-4 rounded-full bg-primary/70 border border-background shadow" />
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2 text-sm">
                      <span className="font-semibold text-foreground">
                        {userLabel}
                      </span>
                      <span className="text-muted-foreground">
                        {getActionLabel(entry.action)}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {formatTimestamp(entry.timestamp)}
                      </span>
                    </div>
                    {entry.object_repr && (
                      <p className="text-sm text-muted-foreground">
                        {entry.object_repr}
                      </p>
                    )}
                    {changes.length > 0 ? (
                      <ul className="space-y-1 text-sm text-foreground/80">
                        {changes.map((change: any, index: number) => (
                          <li key={`${entry.id}-change-${index}`}>
                            {change}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-muted-foreground">
                        No change details recorded.
                      </p>
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">
            No history events recorded yet.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
