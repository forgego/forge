import React, { useMemo, useState, useEffect, useCallback } from "react";
import { Link, useNavigate, useLocation } from "@tanstack/react-router";
import { useModels, useConfig, useLogout } from "../../api/hooks/adminHooks";
import { Button } from "../ui/button";
import {
  LayoutDashboard,
  LogOut,
  Menu,
  Bell,
  Package,
  ChevronDown,
  Star,
  ChevronLeft,
  ChevronRight,
  Keyboard,
  Search,
  X,
} from "lucide-react";
import { cn } from "../../lib/utils";
import { ModelIcon } from "../ModelIcon";

import { GlobalSearch } from "./GlobalSearch";
import { Breadcrumbs } from "./Breadcrumbs";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";
import {
  useShortcutHelp,
  ShortcutHelpDialog,
} from "../../hooks/useKeyboardShortcuts";

const normalizeAdminPath = (path?: string) =>
  path?.startsWith("/admin/") ? path.substring(6) : path;

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { data: modelsData } = useModels();
  const { data: configData } = useConfig();
  const logoutMutation = useLogout();
  const navigate = useNavigate();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCompact, setSidebarCompact] = useState(false);
  const [pinnedModels, setPinnedModels] = useState<string[]>(() => {
    if (typeof window === "undefined") return [];
    const stored = localStorage.getItem("forge.admin.pinnedModels");
    if (!stored) return [];
    try {
      const parsed = JSON.parse(stored);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  });
  const [expandedSections, setExpandedSections] = useState<
    Record<string, boolean>
  >({});
  const [modelFilter, setModelFilter] = useState("");
  const { showHelp, setShowHelp } = useShortcutHelp();

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token && location.pathname !== "/login") {
      navigate({ to: "/login" });
    }
  }, [navigate, location.pathname]);

  // Close the mobile drawer on navigation and on Escape.
  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!sidebarOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setSidebarOpen(false);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [sidebarOpen]);

  const go = useCallback(
    (to: string, params?: Record<string, string>) => {
      setSidebarOpen(false);
      if (params) {
        navigate({ to, params } as any);
      } else {
        navigate({ to } as any);
      }
    },
    [navigate]
  );

  const handleLogout = () => {
    logoutMutation.mutate(undefined, {
      onSettled: () => navigate({ to: "/login" }),
    });
  };

  const isEntryMatch = (entry: any): boolean => {
    const normalizedPath = normalizeAdminPath(entry.path);
    return Boolean(normalizedPath && normalizedPath === location.pathname);
  };

  const isMenuEntryActive = (entry: any): boolean => {
    if (isEntryMatch(entry)) return true;
    return entry.children?.some(isMenuEntryActive) ?? false;
  };

  const findActiveMenuEntry = (entries: any[]): any | null => {
    for (const entry of entries) {
      if (isEntryMatch(entry)) return entry;
      if (entry.children?.length) {
        const nested = findActiveMenuEntry(entry.children);
        if (nested) return nested;
      }
    }
    return null;
  };

  // Recursive Sidebar Item (button-based: keyboard + screen-reader friendly)
  const SidebarItem = ({
    item,
    depth = 0,
    compact = false,
  }: {
    item: any;
    depth?: number;
    compact?: boolean;
  }) => {
    const hasChildren = item.children && item.children.length > 0;
    const [expanded, setExpanded] = useState(false);
    const active = isMenuEntryActive(item);

    // Auto-expand if active child
    useEffect(() => {
      if (hasChildren && active) {
        setExpanded(true);
      }
    }, [location.pathname, hasChildren, active]);

    const handleClick = () => {
      if (hasChildren) {
        setExpanded(!expanded);
      } else {
        const path = normalizeAdminPath(item.path);
        if (path) {
          go(path);
        }
      }
    };

    return (
      <div>
        <button
          type="button"
          onClick={handleClick}
          aria-current={isEntryMatch(item) ? "page" : undefined}
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
              className="h-6 w-6 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-[10px] font-semibold text-muted-foreground group-hover:text-foreground"
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
              />
            ))}
          </div>
        )}
      </div>
    );
  };

  const models = modelsData?.models || [];
  const plugins = configData?.plugins || [];
  const sessionUser = (configData as any)?.user as
    | { name?: string; username?: string; role?: string }
    | undefined;
  const userName =
    sessionUser?.name || sessionUser?.username || "Admin";
  const userRole = sessionUser?.role || "Super Admin";
  const userInitials = userName
    .split(/[\s_.-]+/)
    .map((part) => part.charAt(0))
    .join("")
    .slice(0, 2)
    .toUpperCase();
  const modelByName = useMemo(
    () => new Map(models.map((model: any) => [model.name, model])),
    [models]
  );

  useEffect(() => {
    localStorage.setItem(
      "forge.admin.pinnedModels",
      JSON.stringify(pinnedModels)
    );
  }, [pinnedModels]);

  const togglePinnedModel = (modelName: string) => {
    setPinnedModels((prev) =>
      prev.includes(modelName)
        ? prev.filter((name) => name !== modelName)
        : [...prev, modelName]
    );
  };

  const formatGroupLabel = (value: string) =>
    value
      .replace(/[_-]+/g, " ")
      .replace(/\b\w/g, (char) => char.toUpperCase());

  const groupByModel = useMemo(() => {
    const configGroups =
      configData?.model_groups ??
      configData?.modelGroups ??
      configData?.model_groups_by_app ??
      null;
    const modelGroupMap = new Map<string, string>();

    if (Array.isArray(configGroups)) {
      configGroups.forEach((group: any) => {
        const label = group.label || group.name || group.app || "Other";
        const groupModels: string[] =
          group.models || group.model_names || group.modelNames || [];
        groupModels.forEach((modelName: string) => {
          modelGroupMap.set(modelName, label);
        });
      });
    } else if (configGroups && typeof configGroups === "object") {
      Object.entries(configGroups).forEach(([label, groupModels]) => {
        if (Array.isArray(groupModels)) {
          groupModels.forEach((modelName: string) => {
            modelGroupMap.set(modelName, label);
          });
        }
      });
    }

    const grouped = new Map<string, any[]>();
    const query = modelFilter.trim().toLowerCase();
    models
      .filter(
        (model: any) =>
          query === "" ||
          model.verbose_name_plural?.toLowerCase().includes(query) ||
          model.name?.toLowerCase().includes(query)
      )
      .forEach((model: any) => {
      const configLabel = modelGroupMap.get(model.name);
      const derivedGroup = (() => {
        if (model.name.includes(".")) {
          return model.name.split(".")[0];
        }
        if (model.name.includes("__")) {
          return model.name.split("__")[0];
        }
        if (model.name.includes("_")) {
          return model.name.split("_")[0];
        }
        return "Other";
      })();
      const label = formatGroupLabel(configLabel || derivedGroup);
      const existing = grouped.get(label) || [];
      existing.push(model);
      grouped.set(label, existing);
    });

    return Array.from(grouped.entries())
      .map(([label, groupModels]) => ({
        label,
        models: groupModels.sort((a: any, b: any) =>
          a.verbose_name_plural.localeCompare(b.verbose_name_plural)
        ),
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [configData, models, modelFilter]);

  const pluginSections = useMemo(
    () =>
      plugins
        .filter((plugin: any) => plugin.menuEntries?.length)
        .map((plugin: any) => ({
          id: `plugin-${plugin.name}`,
          label: plugin.label || plugin.name,
          entries: plugin.menuEntries,
        })),
    [plugins]
  );

  const activeModelSectionId = useMemo(() => {
    const activeModel = models.find((model: any) =>
      location.pathname.startsWith(`/${model.name}`)
    );
    if (!activeModel) return null;
    let groupKey: string;
    if (activeModel.name.includes(".")) {
      groupKey = activeModel.name.split(".")[0];
    } else if (activeModel.name.includes("_")) {
      groupKey = activeModel.name.split("_")[0];
    } else {
      groupKey = activeModel.name.charAt(0).toUpperCase();
    }
    return `models-${groupKey.toLowerCase()}`;
  }, [location.pathname, models]);

  const activePluginSectionId = useMemo(() => {
    const activePlugin = plugins.find((plugin: any) =>
      plugin.menuEntries?.some(isMenuEntryActive)
    );
    return activePlugin ? `plugin-${activePlugin.name}` : null;
  }, [location.pathname, plugins]);

  const activePluginInfo = useMemo(() => {
    for (const plugin of plugins) {
      const entries = plugin.menuEntries ?? [];
      const activeEntry = findActiveMenuEntry(entries);
      if (activeEntry) {
        return { plugin, entry: activeEntry };
      }
    }
    return { plugin: null, entry: null };
  }, [plugins, location.pathname]);

  useEffect(() => {
    setExpandedSections((prev) => ({
      ...prev,
      ...(activeModelSectionId ? { [activeModelSectionId]: true } : {}),
      ...(activePluginSectionId ? { [activePluginSectionId]: true } : {}),
    }));
  }, [activeModelSectionId, activePluginSectionId]);

  const showPluginHeader = Boolean(activePluginInfo.plugin);
  const pluginLabel =
    activePluginInfo.plugin?.label || activePluginInfo.plugin?.name;
  const entryLabel = activePluginInfo.entry?.label;

  return (
    <div className="min-h-screen bg-background flex">
      {/* Mobile backdrop */}
      {sidebarOpen && (
        <button
          type="button"
          aria-label="Close navigation"
          onClick={() => setSidebarOpen(false)}
          className="fixed inset-0 z-40 bg-background/60 backdrop-blur-sm lg:hidden"
        />
      )}
      {/* Sidebar */}
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
          {/* Main Nav */}
          <div className="space-y-1">
            <Link
              to="/"
              data-testid="nav-dashboard"
              aria-current={location.pathname === "/" ? "page" : undefined}
              onClick={() => setSidebarOpen(false)}
              className={cn(
                "relative flex items-center gap-2.5 px-3 py-2 rounded-lg transition-all hover:bg-accent hover:text-accent-foreground group mb-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                location.pathname === "/" &&
                  "bg-primary/[0.08] hover:bg-primary/[0.12] font-medium",
                sidebarCompact && "justify-center px-2"
              )}
              title={sidebarCompact ? "Dashboard" : undefined}
            >
              {location.pathname === "/" && !sidebarCompact && (
                <span
                  aria-hidden
                  className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-full bg-primary"
                />
              )}
              <span
                aria-hidden
                className={cn(
                  "flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border transition-colors",
                  location.pathname === "/"
                    ? "border-primary/30 bg-primary/10 text-primary"
                    : "border-border/60 bg-muted/50 text-muted-foreground group-hover:text-foreground"
                )}
              >
                <LayoutDashboard className="h-3.5 w-3.5" />
              </span>
              {!sidebarCompact && (
                <span className="text-[13px]">Dashboard</span>
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
                const active = location.pathname.startsWith(`/${model.name}`);
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
                      className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-amber-500/30 bg-amber-500/10 text-amber-500"
                    >
                      <Star className="h-3.5 w-3.5" fill="currentColor" />
                    </span>
                    {!sidebarCompact && (
                      <span className="text-[13px] truncate">
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
                  className="h-9 w-full rounded-lg border border-border/60 bg-background pl-8 pr-7 text-[13px] outline-none placeholder:text-muted-foreground focus:border-primary/40 focus:ring-2 focus:ring-ring"
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
                {!sidebarCompact && (
                  <h5 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                    {group.label}
                  </h5>
                )}
                {group.models.map((model: any) => {
                  const active = location.pathname.startsWith(`/${model.name}`);
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
                        {active && !sidebarCompact && (
                          <span
                            aria-hidden
                            className="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-full bg-primary"
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
                            <span className="text-[13px] flex-1 truncate">
                              {model.verbose_name_plural}
                            </span>
                            {typeof model.count === "number" && (
                              <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground/80">
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
                            pinned && "opacity-100 text-yellow-500"
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
                          />
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </nav>

        <div className="p-4 border-t border-border shrink-0">
          <Button
            variant="ghost"
            data-testid="logout-button"
            className={cn(
              "w-full text-muted-foreground hover:text-destructive hover:bg-destructive/10",
              sidebarCompact ? "justify-center px-2" : "justify-start"
            )}
            onClick={handleLogout}
            disabled={logoutMutation.isPending}
            title={sidebarCompact ? "Logout" : undefined}
          >
            <LogOut className={cn("h-4 w-4", !sidebarCompact && "mr-3")} aria-hidden />
            {!sidebarCompact && (
              <span className="text-sm font-medium">Logout</span>
            )}
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-h-screen min-w-0 bg-muted/20">
        <header className="h-16 border-b border-border/50 flex items-center px-4 sm:px-6 bg-card/80 backdrop-blur-md sticky top-0 z-30 shrink-0 gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden text-muted-foreground shrink-0"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open navigation"
          >
            <Menu className="h-5 w-5" />
          </Button>

          <div className="min-w-0 flex-1 max-w-md">
            <GlobalSearch models={models} />
          </div>

          {showPluginHeader && (
            <nav
              aria-label="Breadcrumb"
              className="hidden xl:flex items-center gap-1.5 text-xs text-muted-foreground min-w-0"
            >
              <span className="font-semibold uppercase tracking-wider">Plugin</span>
              <ChevronRight className="h-3 w-3 shrink-0" aria-hidden />
              <span className="text-foreground text-sm font-medium truncate">
                {pluginLabel}
              </span>
              {entryLabel && (
                <>
                  <ChevronRight className="h-3 w-3 shrink-0" aria-hidden />
                  <span className="text-sm truncate">{entryLabel}</span>
                </>
              )}
            </nav>
          )}

          <div className="ml-auto flex items-center gap-1 sm:gap-2 shrink-0">
            <Button
              variant="ghost"
              size="icon"
              className="text-muted-foreground hover:text-foreground"
              onClick={() => setShowHelp(true)}
              title="Keyboard Shortcuts (?)"
              aria-label="Keyboard Shortcuts"
            >
              <Keyboard className="h-4 w-4" />
            </Button>
            <ThemeCustomizer />
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full text-muted-foreground hover:text-foreground"
              title="Notifications"
              aria-label="Notifications (none)"
            >
              <Bell className="h-4 w-4" />
            </Button>
            <div className="h-6 w-[1px] bg-border mx-1 sm:mx-2" aria-hidden />
            <div className="flex items-center gap-3 pl-1 sm:pl-2">
              <div className="text-right hidden md:block">
                <p className="text-sm font-semibold leading-none text-foreground">
                  {userName}
                </p>
                <p className="text-[10px] text-muted-foreground mt-1 uppercase font-bold tracking-wider">
                  {userRole}
                </p>
              </div>
              <div
                className="w-8 h-8 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold"
                title={`${userName} (${userRole})`}
                aria-hidden
              >
                {userInitials}
              </div>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <div className="flex-1 p-4 sm:p-6 lg:p-8 overflow-auto">
          <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">
            <Breadcrumbs />
            {children}
          </div>
        </div>
        <ShortcutHelpDialog open={showHelp} onOpenChange={setShowHelp} />
      </main>
    </div>
  );
}
