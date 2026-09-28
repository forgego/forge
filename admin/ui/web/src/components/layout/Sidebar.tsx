import { useMemo, type Dispatch, type SetStateAction } from "react";
import { motion, LayoutGroup } from "framer-motion";
import { Link } from "@tanstack/react-router";
import { Button } from "../ui/button";
import { LayoutDashboard, LogOut, Package, ChevronDown, Star, ChevronLeft, ChevronRight, Search, X } from "lucide-react";
import { cn } from "../../lib/utils";
import { ModelIcon } from "../ModelIcon";
import { SidebarItem } from "./SidebarItem";

export interface SidebarProps {
  models: any[];
  groupByModel: { label: string; models: any[] }[];
  pluginSections: { id: string; label: string; entries: any[] }[];
  pinnedModels: string[]; togglePinnedModel: (modelName: string) => void;
  modelFilter: string; setModelFilter: Dispatch<SetStateAction<string>>;
  expandedSections: Record<string, boolean>; setExpandedSections: Dispatch<SetStateAction<Record<string, boolean>>>;
  sidebarOpen: boolean; setSidebarOpen: Dispatch<SetStateAction<boolean>>;
  sidebarCompact: boolean; setSidebarCompact: Dispatch<SetStateAction<boolean>>;
  pathname: string; go: (to: string, params?: Record<string, string>) => void;
  onLogout: () => void; logoutPending: boolean; prefersReducedMotion: boolean | null;
}

export function Sidebar({
  models, groupByModel, pluginSections, pinnedModels, togglePinnedModel,
  modelFilter, setModelFilter, expandedSections, setExpandedSections,
  sidebarOpen, setSidebarOpen, sidebarCompact, setSidebarCompact,
  pathname, go, onLogout, logoutPending, prefersReducedMotion,
}: SidebarProps) {
  const modelByName = useMemo(() => new Map(models.map((model: any) => [model.name, model])), [models]);
  const isDashboardActive = pathname === "/";
  return (
      <aside
        aria-label="Admin navigation"
        className={cn(
          "fixed inset-y-0 left-0 z-50 bg-card border-r border-border transition-all duration-300 ease-in-out lg:static lg:translate-x-0 flex flex-col",
          sidebarCompact ? "lg:w-20" : "lg:w-72",
          sidebarOpen ? "translate-x-0 w-72" : "-translate-x-full lg:translate-x-0"
        )}
      >
        <div className="h-16 flex items-center px-4 border-b border-border/50 shrink-0">
          <div className="flex items-center gap-2.5 flex-1 min-w-0">
            <div className="w-8 h-8 shrink-0 rounded-lg bg-primary flex items-center justify-center text-primary-foreground font-bold shadow-sm">
              F
            </div>
            {!sidebarCompact && (
              <span className="text-lg font-bold tracking-tight text-foreground truncate">
                Forge Admin
              </span>
            )}
          </div>
          {!sidebarCompact && (
            <Button
              variant="ghost"
              size="icon"
              className="lg:hidden text-muted-foreground"
              onClick={() => setSidebarOpen(false)}
              aria-label="Close navigation"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="hidden lg:inline-flex text-muted-foreground"
            onClick={() => setSidebarCompact((prev) => !prev)}
            aria-label={sidebarCompact ? "Expand sidebar" : "Collapse sidebar"}
          >
            {sidebarCompact ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>
        <nav className="p-4 space-y-6 overflow-y-auto flex-1" aria-label="Primary">
          <LayoutGroup id="admin-nav">
          {/* Main Nav */}
          <div className="space-y-1">
            <Link
              to="/"
              data-testid="nav-dashboard"
              aria-current={pathname === "/" ? "page" : undefined}
              onClick={() => setSidebarOpen(false)}
              className={cn(
                "relative flex items-center gap-2.5 px-3 py-2 rounded-lg transition-all hover:bg-accent hover:text-accent-foreground group mb-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                pathname === "/" &&
                  "bg-primary/[0.08] hover:bg-primary/[0.12] font-medium",
                sidebarCompact && "justify-center px-2"
              )}
              title={sidebarCompact ? "Dashboard" : undefined}
            >
              {isDashboardActive && !sidebarCompact && (
                <motion.span
                  aria-hidden
                  layoutId="admin-nav-indicator"
                  className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-full bg-primary"
                  transition={
                    prefersReducedMotion
                      ? { duration: 0 }
                      : { type: "spring", stiffness: 520, damping: 38 }
                  }
                />
              )}
              <span
                aria-hidden
                className={cn(
                  "flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border transition-colors",
                  pathname === "/"
                    ? "border-primary/30 bg-primary/10 text-primary"
                    : "border-border/60 bg-muted/50 text-muted-foreground group-hover:text-foreground"
                )}
              >
                <LayoutDashboard className="h-3.5 w-3.5" />
              </span>
              {!sidebarCompact && (
                <span className="text-body">Dashboard</span>
              )}
            </Link>
          </div>

          {/* Pinned Models */}
          {pinnedModels.length > 0 && (
            <div className="space-y-1">
              {!sidebarCompact && (
                <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                  Pinned
                </h4>
              )}
              {pinnedModels.map((modelName) => {
                const model = modelByName.get(modelName);
                if (!model) return null;
                const active = pathname.startsWith(`/${model.name}`);
                return (
                  <button
                    key={model.name}
                    type="button"
                    data-testid={`nav-pinned-${model.name}`}
                    onClick={() => go("/$model", { model: model.name })}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex w-full items-center gap-2.5 px-3 py-2 rounded-lg transition-all hover:bg-accent hover:text-accent-foreground group mb-0.5 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                      active && "bg-primary/[0.08] text-foreground font-medium hover:bg-primary/[0.12]",
                      sidebarCompact && "justify-center px-2"
                    )}
                    title={sidebarCompact ? model.verbose_name_plural : undefined}
                  >
                    <span
                      aria-hidden
                      className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-warning/20 bg-warning-surface text-warning"
                    >
                      <Star className="h-3.5 w-3.5" fill="currentColor" />
                    </span>
                    {!sidebarCompact && (
                      <span className="text-body truncate">
                        {model.verbose_name_plural}
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          )}

          {/* Models Section */}
          <div className="space-y-4">
            {!sidebarCompact && (
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3">
                Content Models
              </h4>
            )}
            {!sidebarCompact && (
              <div className="relative px-0">
                <label htmlFor="sidebar-model-filter" className="sr-only">
                  Filter models
                </label>
                <Search
                  aria-hidden
                  className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
                />
                <input
                  id="sidebar-model-filter"
                  value={modelFilter}
                  onChange={(e) => setModelFilter(e.target.value)}
                  placeholder="Filter models…"
                  autoComplete="off"
                  className="h-9 w-full rounded-lg border border-border/60 bg-background pl-8 pr-7 text-body outline-none placeholder:text-muted-foreground focus:border-primary/40 focus:ring-2 focus:ring-ring"
                />
                {modelFilter && (
                  <button
                    type="button"
                    aria-label="Clear model filter"
                    onClick={() => setModelFilter("")}
                    className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:text-foreground"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            )}
            {groupByModel.length === 0 && (
              <p className="px-3 py-4 text-xs text-muted-foreground">
                {modelFilter
                  ? `No models match “${modelFilter}”.`
                  : "No models registered."}
              </p>
            )}
{groupByModel.map((group) => (
  <div key={group.label} className="space-y-1">
    {!sidebarCompact && group.label !== "" && groupByModel.length > 1 && (
      <h5 className="text-micro font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
        {group.label}
      </h5>
    )}
    {group.models.map((model: any) => {
                  const active = pathname.startsWith(`/${model.name}`);
                  const pinned = pinnedModels.includes(model.name);
                  return (
                    <div
                      key={model.name}
                      className="group relative"
                    >
                      <button
                        type="button"
                        data-testid={`nav-${model.name}`}
                        onClick={() => go("/$model", { model: model.name })}
                        aria-current={active ? "page" : undefined}
                        className={cn(
                          "relative flex w-full items-center gap-2.5 px-3 py-2 rounded-lg transition-all hover:bg-accent hover:text-accent-foreground mb-0.5 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring group",
                          active && "bg-primary/[0.08] text-foreground font-medium hover:bg-primary/[0.12]",
                          sidebarCompact && "justify-center px-2"
                        )}
                        title={
                          sidebarCompact ? model.verbose_name_plural : undefined
                        }
                      >
                        {active && !isDashboardActive && !sidebarCompact && (
                          <motion.span
                            aria-hidden
                            layoutId="admin-nav-indicator"
                            className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-full bg-primary"
                            transition={
                              prefersReducedMotion
                                ? { duration: 0 }
                                : { type: "spring", stiffness: 520, damping: 38 }
                            }
                          />
                        )}
                        {sidebarCompact ? (
                          <span
                            aria-hidden
                            className={cn(
                              "flex h-8 w-8 items-center justify-center rounded-lg border transition-colors",
                              active
                                ? "border-primary/30 bg-primary/10 text-primary"
                                : "border-border/60 bg-muted/60 text-muted-foreground group-hover:text-foreground"
                            )}
                          >
                            <ModelIcon name={model.icon} className="h-4 w-4" />
                          </span>
                        ) : (
                          <>
                            <span
                              aria-hidden
                              className={cn(
                                "flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border transition-colors",
                                active
                                  ? "border-primary/30 bg-primary/10 text-primary"
                                  : "border-border/60 bg-muted/50 text-muted-foreground group-hover:text-foreground"
                              )}
                            >
                              <ModelIcon name={model.icon} className="h-3.5 w-3.5" />
                            </span>
                            <span className="text-body flex-1 truncate">
                              {model.verbose_name_plural}
                            </span>
                            {typeof model.count === "number" && (
                              <span className="shrink-0 text-micro tabular-nums text-muted-foreground/80">
                                {model.count > 999
                                  ? `${(model.count / 1000).toFixed(1)}k`
                                  : model.count}
                              </span>
                            )}
                          </>
                        )}
                      </button>
                      {!sidebarCompact && (
                        <button
                          type="button"
                          aria-label={
                            pinned
                              ? `Unpin ${model.verbose_name_plural}`
                              : `Pin ${model.verbose_name_plural}`
                          }
                          aria-pressed={pinned}
                          onClick={(event) => {
                            event.stopPropagation();
                            togglePinnedModel(model.name);
                          }}
                          className={cn(
                            "absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-1.5 text-muted-foreground hover:text-primary hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring opacity-100 lg:opacity-0 lg:group-hover:opacity-100 lg:group-focus-within:opacity-100 focus-visible:opacity-100",
                            pinned && "opacity-100 text-warning"
                          )}
                        >
                          <Star
                            className="h-4 w-4"
                            fill={pinned ? "currentColor" : "none"}
                            aria-hidden
                          />
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            ))}
          </div>

          {/* Plugins Section */}
          {pluginSections.length > 0 && (
            <div className="space-y-1">
              {!sidebarCompact && (
                <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                  Plugins
                </h4>
              )}
              {pluginSections.map((section: any) => {
                const isExpanded = expandedSections[section.id] ?? false;
                return (
                  <div key={section.id} className="space-y-1">
                    <button
                      type="button"
                      onClick={() =>
                        setExpandedSections((prev) => ({
                          ...prev,
                          [section.id]: !isExpanded,
                        }))
                      }
                      className={cn(
                        "flex items-center w-full gap-2 px-3 py-2 rounded-md text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground hover:bg-accent/40 transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                        sidebarCompact && "justify-center px-2"
                      )}
                      title={sidebarCompact ? section.label : undefined}
                      aria-expanded={isExpanded}
                    >
                      <Package className="h-3.5 w-3.5 shrink-0" aria-hidden />
                      {!sidebarCompact && <span className="truncate">{section.label}</span>}
                      {!sidebarCompact && (
                        <ChevronDown
                          aria-hidden
                          className={cn(
                            "ml-auto h-3 w-3 transition-transform shrink-0",
                            isExpanded && "rotate-180"
                          )}
                        />
                      )}
                    </button>
                    {isExpanded && (
                      <div className="space-y-1">
                        {section.entries.map((entry: any, idx: number) => (
                          <SidebarItem
                            key={`${section.id}-${idx}`}
                            item={entry}
                            compact={sidebarCompact}
                            pathname={pathname}
                            onNavigate={go}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
          </LayoutGroup>
        </nav>

        <div className="p-4 border-t border-border shrink-0">
          <Button
            variant="ghost"
            data-testid="logout-button"
            className={cn(
              "w-full text-muted-foreground hover:text-destructive hover:bg-destructive/10",
              sidebarCompact ? "justify-center px-2" : "justify-start"
            )}
            onClick={onLogout}
            disabled={logoutPending}
            title={sidebarCompact ? "Logout" : undefined}
          >
            <LogOut className={cn("h-4 w-4", !sidebarCompact && "mr-3")} aria-hidden />
            {!sidebarCompact && (
              <span className="text-sm font-medium">Logout</span>
            )}
          </Button>
        </div>
      </aside>
  );
}
