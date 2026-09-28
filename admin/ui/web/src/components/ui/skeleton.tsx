import { cn } from "@/lib/utils";

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "text" | "circular" | "rectangular";
  width?: string | number;
  height?: string | number;
}

export function Skeleton({
  className,
  variant = "text",
  width,
  height,
  ...props
}: SkeletonProps) {
  const variantClasses = {
    text: "rounded-md",
    circular: "rounded-full",
    rectangular: "rounded-lg",
  };

  const style: React.CSSProperties = {
    width: width,
    height: height || (variant === "text" ? "1em" : undefined),
  };

  return (
    <div
      className={cn(
        "animate-pulse bg-muted",
        variantClasses[variant],
        className
      )}
      style={style}
      {...props}
    />
  );
}

export function SkeletonCard() {
  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm">
      <div className="flex items-center space-x-4">
        <Skeleton variant="circular" width={48} height={48} />
        <div className="space-y-2 flex-1">
          <Skeleton width="60%" height={16} />
          <Skeleton width="40%" height={12} />
        </div>
      </div>
      <div className="space-y-3 mt-4">
        <Skeleton width="100%" height={14} />
        <Skeleton width="80%" height={14} />
        <Skeleton width="90%" height={14} />
      </div>
    </div>
  );
}

export function SkeletonTable({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="rounded-md border">
      <div className="border-b">
        <div className="flex">
          {Array.from({ length: cols }).map((_, i) => (
            <div key={i} className="flex-1 p-4">
              <Skeleton width="80%" height={14} />
            </div>
          ))}
        </div>
      </div>
      <div>
        {Array.from({ length: rows }).map((_, rowI) => (
          <div key={rowI} className="flex border-b last:border-0">
            {Array.from({ length: cols }).map((_, colI) => (
              <div key={colI} className="flex-1 p-4">
                <Skeleton width={colI === 0 ? "60%" : "80%"} height={12} />
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

export function SkeletonForm({ fields = 4 }: { fields?: number }) {
  return (
    <div className="space-y-6">
      {Array.from({ length: fields }).map((_, i) => (
        <div key={i} className="space-y-2">
          <Skeleton width={120} height={14} />
          <Skeleton width="100%" height={40} />
        </div>
      ))}
    </div>
  );
}

export function SkeletonList({ items = 4 }: { items?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} className="flex items-center space-x-4 p-4 border rounded-lg">
          <Skeleton variant="circular" width={40} height={40} />
          <div className="space-y-2 flex-1">
            <Skeleton width="40%" height={16} />
            <Skeleton width="60%" height={12} />
          </div>
        </div>
      ))}
    </div>
  );
}
