import { useState, useEffect } from "react";
import { Package, ChevronDown } from "lucide-react";
import { cn } from "../../lib/utils";
import { isEntryMatch, isMenuEntryActive, normalizeAdminPath } from "./nav-utils";

export interface SidebarItemProps {
  item: any;
  depth?: number;
  compact?: boolean;
  pathname: string;
  onNavigate: (path: string) => void;
}

export const SidebarItem = ({
  item,
  depth = 0,
  compact = false,
  pathname,
  onNavigate,
}: SidebarItemProps) => {
  const hasChildren = item.children && item.children.length > 0;
  const [expanded, setExpanded] = useState(false);
  const active = isMenuEntryActive(item, pathname);

  // Auto-expand if active child
  useEffect(() => {
    if (hasChildren && active) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- sync expanded state with pathname changes
      setExpanded(true);
    }
  }, [pathname, hasChildren, active]);

  const handleClick = () => {
    if (hasChildren) {
      setExpanded(!expanded);
    } else {
      const path = normalizeAdminPath(item.path);
      if (path) {
        onNavigate(path);
      }
    }
  };

  return (
    <div>
      <button
        type="button"
        onClick={handleClick}
        aria-current={isEntryMatch(item, pathname) ? "page" : undefined}
        aria-expanded={hasChildren ? expanded : undefined}
        className={cn(
          "flex w-full items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground text-sm text-left group select-none mb-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          active && "bg-accent text-accent-foreground font-medium",
          depth > 0 && "text-muted-foreground",
          compact && "justify-center px-2"
        )}
        style={{
          paddingLeft:
            depth === 0 || compact
              ? "0.75rem"
              : `${depth * 1 + 0.75}rem`,
        }}
        title={compact ? item.label : undefined}
      >
        {depth === 0 && (
          <Package className="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0" aria-hidden />
        )}
        {compact && depth > 0 ? (
          <span
            aria-hidden
            className="h-6 w-6 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-micro font-semibold text-muted-foreground group-hover:text-foreground"
          >
            {item.label?.charAt(0).toUpperCase()}
          </span>
        ) : (
          <span className="flex-1 truncate">{item.label}</span>
        )}
        {hasChildren && !compact && (
          <ChevronDown
            aria-hidden
            className={cn(
              "h-3 w-3 transition-transform shrink-0 text-muted-foreground",
              expanded && "rotate-180"
            )}
          />
        )}
      </button>
      {hasChildren && expanded && !compact && (
        <div className="space-y-1 pt-1">
          {item.children.map((child: any, idx: number) => (
            <SidebarItem
              key={idx}
              item={child}
              depth={depth + 1}
              compact={compact}
              pathname={pathname}
              onNavigate={onNavigate}
            />
          ))}
        </div>
      )}
    </div>
  );
};
