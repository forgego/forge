import { AlertTriangle, X } from "lucide-react";
import { Button } from "../ui/button";
import type { BulkResult } from "./bulk-result";

/**
 * Persistent per-item report for a bulk operation that did not apply to
 * every selected record. A toast alone disappears before an operator can
 * see which records were skipped and why.
 */
export function BulkResultPanel({
  result,
  onDismiss,
}: {
  result: BulkResult;
  onDismiss: () => void;
}) {
  const total = result.succeeded + result.failures.length;
  return (
    <div
      role="alert"
      data-testid="bulk-result"
      className="rounded-lg border border-warning/20 bg-warning-surface px-4 py-3 text-body text-foreground"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-2">
          <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0 text-warning" aria-hidden />
          <p className="font-medium" data-testid="bulk-result-summary">
            {result.label}: applied to{" "}
            <span className="font-mono tabular-nums">{result.succeeded}</span> of{" "}
            <span className="font-mono tabular-nums">{total}</span> selected;{" "}
            <span className="font-mono tabular-nums">{result.failures.length}</span>{" "}
            {result.failures.length === 1 ? "record was" : "records were"} not changed.
          </p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 p-0 text-muted-foreground"
          aria-label="Dismiss bulk result"
          data-testid="bulk-result-dismiss"
          onClick={onDismiss}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
      <ul className="mt-2 space-y-1 pl-6 text-meta">
        {result.failures.map((failure, idx) => (
          <li key={`${failure.id}-${idx}`} data-testid={`bulk-result-item-${failure.id}`}>
            <span className="font-mono tabular-nums">#{String(failure.id)}</span>{" "}
            — {failure.message}
          </li>
        ))}
      </ul>
    </div>
  );
}
