import { useState } from "react";
import { AlertTriangle, X } from "lucide-react";
import { getExtensionFailures, type ExtensionFailure } from "./loader";
import { cn } from "../utils";

export interface ExtensionErrorBannerProps {
  failures?: ExtensionFailure[];
  onDismiss?: () => void;
  className?: string;
}

export function ExtensionErrorBanner({
  failures: propFailures,
  onDismiss,
  className,
}: ExtensionErrorBannerProps) {
  const [dismissed, setDismissed] = useState(false);
  const failures = propFailures ?? getExtensionFailures();

  if (dismissed || !failures || failures.length === 0) {
    return null;
  }

  const failedNames = Array.from(new Set(failures.map((f) => f.name)));

  const handleDismiss = () => {
    setDismissed(true);
    onDismiss?.();
  };

  return (
    <div
      role="alert"
      data-testid="extension-error-banner"
      className={cn(
        "flex items-start justify-between gap-3 p-4 rounded-lg border border-warning/30 bg-warning-surface text-foreground text-sm shadow-sm transition-all",
        className
      )}
    >
      <div className="flex items-start gap-3">
        <AlertTriangle className="h-5 w-5 text-warning shrink-0 mt-0.5" />
        <div className="space-y-1">
          <p className="font-semibold text-foreground">
            Some extensions failed to load
          </p>
          <p className="text-xs text-muted-foreground leading-relaxed">
            The following extension{failedNames.length > 1 ? "s" : ""} could not be initialized:{" "}
            <span className="font-medium text-foreground">
              {failedNames.join(", ")}
            </span>
            . The rest of the admin interface remains fully functional.
          </p>
        </div>
      </div>
      <button
        type="button"
        onClick={handleDismiss}
        className="text-muted-foreground hover:text-foreground p-1 rounded-md transition-colors shrink-0"
        aria-label="Dismiss extension error banner"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}

export default ExtensionErrorBanner;
