import React, { useMemo, useState, useEffect } from "react";
import { Link, useNavigate, useLocation } from "@tanstack/react-router";
import { useModels, useConfig } from "../../api/hooks/adminHooks";
import { Button } from "../ui/button";
import {
  LayoutDashboard,
  LogOut,
  Menu,
  Database,
  Bell,
  Package,
  ChevronDown,
  PanelLeftOpen,
  Star,
  ChevronLeft,
  ChevronRight,
  Plus,
} from "lucide-react";
import { cn } from "../../lib/utils";

import { GlobalSearch } from "./GlobalSearch";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";

<<<<<<< codex/add-quick-actions-to-modellistpage
type QuickAction = {
  label: string;
  onClick: () => void;
  disabled?: boolean;
  hidden?: boolean;
  icon?: React.ReactNode;
  ariaLabel?: string;
};
=======
const buildSectionId = (label: string) =>
  `models-${label
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "")}`;
const normalizeAdminPath = (path?: string) =>
  path?.startsWith("/admin/") ? path.substring(6) : path;

const isExternalIcon = (icon?: string) =>
  Boolean(icon && (icon.startsWith("http") || icon.startsWith("/") || icon.startsWith("data:")));
>>>>>>> v8

export default function AdminLayout({
  children,
  quickActions = [],
}: {
  children: React.ReactNode;
  quickActions?: QuickAction[];
}) {
  const { data: modelsData } = useModels();
  const { data: configData } = useConfig();
  const navigate = useNavigate();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [sidebarCompact, setSidebarCompact] = useState(true);
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

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
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

  const renderIconBadge = (icon?: string, label?: string) => {
    const fallbackText = (label ?? icon ?? "?").charAt(0).toUpperCase();
    return (
      <div className="h-8 w-8 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-[10px] font-semibold text-muted-foreground">
        {isExternalIcon(icon) ? (
          <img
            src={icon}
            alt={label ?? "Plugin icon"}
            className="h-4 w-4 object-contain"
          />
        ) : (
          fallbackText
        )}
      </div>
    );
  };

  // Recursive Sidebar Item
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

    // Auto-expand if active child
    useEffect(() => {
      if (hasChildren && isMenuEntryActive(item)) {
        setExpanded(true);
      }
    }, [location.pathname, item, hasChildren]);

    const handleActivate = () => {
      if (hasChildren) {
        setExpanded((prev) => !prev);
        return;
        setExpanded(!expanded);
      } else {
        const path = normalizeAdminPath(item.path);
        if (path) {
          navigate({ to: path });
        }
      }
      const path = item.path.startsWith("/admin/")
        ? item.path.substring(6)
        : item.path;
      navigate({ to: path });
    };

    return (
      <div>
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            handleActivate();
          }}
          className={cn(
            "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer text-sm group select-none mb-1 w-full text-left",
            location.pathname === item.path &&
            "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer text-sm group select-none mb-1",
            isMenuEntryActive(item) &&
              "bg-sidebar-accent text-sidebar-accent-foreground font-medium",
            depth > 0 && "text-muted-foreground",
            compact && "justify-center px-2"
          )}
          style={{
            paddingLeft:
              depth === 0 || compact ? "0.75rem" : `${depth * 1 + 0.75}rem`,
              depth === 0 || compact
                ? "0.75rem"
                : `${depth * 1 + 0.75}rem`,
          }}
          title={compact ? item.label : undefined}
          aria-label={compact ? item.label : undefined}
          aria-expanded={hasChildren ? expanded : undefined}
        >
          {depth === 0 && (
            <Package className="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0" />
          )}
          {compact && depth > 0 && (
          {compact && depth > 0 ? (
            <div className="h-6 w-6 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-[10px] font-semibold text-muted-foreground group-hover:text-foreground">
              {item.label?.charAt(0).toUpperCase()}
            </div>
          ) : (
            <span className="flex-1 truncate">{item.label}</span>
          )}
          {!compact && <span className="flex-1 truncate">{item.label}</span>}
          {hasChildren && !compact && (
            <ChevronDown
              className={cn(
                "h-3 w-3 transition-transform shrink-0 text-muted-foreground",
                expanded && "rotate-180"
              )}
            />
          )}
        </button>
        </div>
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

  const modelGroups = useMemo(() => {
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

    const grouped = new Map<
      string,
      { id: string; label: string; models: any[] }
    >();
    models.forEach((model: any) => {
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
      if (!grouped.has(label)) {
        grouped.set(label, {
          id: buildSectionId(label),
          label,
          models: [],
        });
      }
      grouped.get(label)?.models.push(model);
    });

    return Array.from(grouped.values())
      .map((group) => ({
        ...group,
        models: group.models.sort((a: any, b: any) =>
          a.verbose_name_plural.localeCompare(b.verbose_name_plural)
        ),
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [configData, models]);

  const modelGroupLookup = useMemo(() => {
    const lookup = new Map<string, string>();
    modelGroups.forEach((group) => {
      group.models.forEach((model: any) => {
        lookup.set(model.name, group.id);
      });
    });
    return lookup;
  }, [modelGroups]);

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
    return modelGroupLookup.get(activeModel.name) ?? null;
  }, [location.pathname, models, modelGroupLookup]);

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

<<<<<<< codex/add-quick-actions-to-modellistpage
  const visibleQuickActions = useMemo(
    () => quickActions.filter((action) => !action.hidden),
    [quickActions]
  );
=======
  const handleModelRowKeyDown = (
    event: React.KeyboardEvent<HTMLDivElement>,
    modelName: string
  ) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      navigate({
        to: "/$model",
        params: { model: modelName },
      });
    }
  };
  const showPluginHeader = Boolean(activePluginInfo.plugin);
  const pluginLabel =
    activePluginInfo.plugin?.label || activePluginInfo.plugin?.name;
  const entryLabel = activePluginInfo.entry?.label;
  const showIconGroup =
    Boolean(activePluginInfo.plugin?.icon) ||
    Boolean(activePluginInfo.entry?.icon);
>>>>>>> v8

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 bg-card/95 backdrop-blur-sm border-r border-border transition-all duration-300 ease-in-out lg:relative lg:translate-x-0",
          sidebarCompact ? "lg:w-20" : "lg:w-64",
          "w-64",
          !sidebarOpen && "-translate-x-full lg:hidden"
        )}
      >
        <div className="h-16 flex items-center px-6 border-b border-border/50">
          <div className="flex items-center gap-2.5 flex-1">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground font-bold shadow-sm">
              F
            </div>
            {!sidebarCompact && (
              <span className="text-lg font-bold tracking-tight text-foreground">
                Forge Admin
              </span>
            )}
          </div>
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
        <div className="p-4 space-y-6 overflow-y-auto max-h-[calc(100vh-140px)] no-scrollbar">
          <div className="flex items-center justify-between gap-2">
            <GlobalSearch
              compact={sidebarCompact}
              triggerLabel="Command palette"
              className={sidebarCompact ? "" : "w-full"}
            />
            {!sidebarCompact && (
              <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-widest">
                Jump to model
              </span>
            )}
          </div>
          {/* Main Nav */}
          <div className="space-y-1">
            <Link
              to="/"
              data-testid="nav-dashboard"
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground group mb-1",
                location.pathname === "/" &&
                  "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm",
                sidebarCompact && "justify-center px-2"
              )}
              title={sidebarCompact ? "Dashboard" : undefined}
            >
              <LayoutDashboard className="h-4 w-4" />
              {!sidebarCompact && (
                <span className="font-medium text-sm">Dashboard</span>
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
                return (
                  <div
                    key={model.name}
                    data-testid={`nav-pinned-${model.name}`}
                    role="button"
                    tabIndex={0}
                    onClick={() => {
                      navigate({
                        to: "/$model",
                        params: { model: model.name },
                      });
                    }}
                    onKeyDown={(event) =>
                      handleModelRowKeyDown(event, model.name)
                    }
                    className={cn(
                      "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground cursor-pointer group mb-1",
                      location.pathname.startsWith(`/${model.name}`) &&
                        "bg-accent text-accent-foreground font-medium",
                      sidebarCompact && "justify-center px-2"
                    )}
                    title={sidebarCompact ? model.verbose_name_plural : undefined}
                    aria-label={
                      sidebarCompact ? model.verbose_name_plural : undefined
                    }
                  >
                    <Star className="h-4 w-4 text-yellow-500" />
                    {!sidebarCompact && (
                      <span className="text-sm flex-1 truncate">
                        {model.verbose_name_plural}
                      </span>
                    )}
                    {!sidebarCompact && (
                      <button
                        type="button"
                        aria-label={`Unpin ${model.verbose_name_plural}`}
                        onClick={(event) => {
                          event.stopPropagation();
                          togglePinnedModel(model.name);
                        }}
                        className="opacity-0 group-hover:opacity-100 transition text-muted-foreground hover:text-primary"
                      >
                        <Star className="h-4 w-4" />
                      </button>
                  >
                    <Star className="h-4 w-4 text-yellow-500" />
                    {!sidebarCompact && (
                      <span className="text-sm">
                        {model.verbose_name_plural}
                      </span>
                    )}
                  </div>
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
            {modelGroups.map((group) => {
              const isExpanded = expandedSections[group.id] ?? false;
              return (
                <div key={group.id} className="space-y-1">
                  <button
                    type="button"
                    onClick={() =>
                      setExpandedSections((prev) => ({
                        ...prev,
                        [group.id]: !isExpanded,
                      }))
                    }
                    className={cn(
                      "flex items-center w-full gap-2 px-3 py-2 rounded-md text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground hover:bg-accent/40 transition-all",
                      sidebarCompact && "justify-center px-2"
                    )}
                    title={sidebarCompact ? group.label : undefined}
                    aria-expanded={isExpanded}
                    aria-controls={`${group.id}-panel`}
                  >
                    <Database className="h-3.5 w-3.5" />
                    {!sidebarCompact && <span>{group.label}</span>}
                    {!sidebarCompact && (
                      <ChevronDown
            {groupByModel.map((group) => (
              <div key={group.label} className="space-y-1">
                {!sidebarCompact && (
                  <h5 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                    {group.label}
                  </h5>
                )}
                {group.models.map((model: any) => (
                  <div
                    key={model.name}
                    data-testid={`nav-${model.name}`}
                    onClick={() => {
                      navigate({
                        to: "/$model",
                        params: { model: model.name },
                      });
                    }}
                    className={cn(
                      "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground cursor-pointer group mb-1",
                      location.pathname.startsWith(`/${model.name}`) &&
                        "bg-accent text-accent-foreground font-medium",
                      sidebarCompact && "justify-center px-2"
                    )}
                    title={
                      sidebarCompact ? model.verbose_name_plural : undefined
                    }
                  >
                    {sidebarCompact ? (
                      <div className="h-7 w-7 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-xs font-semibold text-muted-foreground group-hover:text-foreground">
                        {model.verbose_name_plural
                          .split(" ")
                          .map((part: string) => part[0])
                          .join("")
                          .slice(0, 2)
                          .toUpperCase()}
                      </div>
                    ) : (
                      <>
                        <Database className="h-4 w-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                        <span className="text-sm flex-1 truncate">
                          {model.verbose_name_plural}
                        </span>
                      </>
                    )}
                    {!sidebarCompact && (
                      <button
                        type="button"
                        aria-label={
                          pinnedModels.includes(model.name)
                            ? `Unpin ${model.verbose_name_plural}`
                            : `Pin ${model.verbose_name_plural}`
                        }
                        onClick={(event) => {
                          event.stopPropagation();
                          togglePinnedModel(model.name);
                        }}
                        className={cn(
                          "ml-auto h-3 w-3 transition-transform",
                          isExpanded && "rotate-180"
                        )}
                      />
                    )}
                  </button>
                  {isExpanded && (
                    <div
                      id={`${group.id}-panel`}
                      className="space-y-1"
                      role="region"
                      aria-label={group.label}
                    >
                      {group.models.map((model: any) => {
                        const canAdd = model.permissions?.add;
                        return (
                          <div
                            key={model.name}
                            data-testid={`nav-${model.name}`}
                            role="button"
                            tabIndex={0}
                            onClick={() => {
                              navigate({
                                to: "/$model",
                                params: { model: model.name },
                              });
                            }}
                            onKeyDown={(event) =>
                              handleModelRowKeyDown(event, model.name)
                            }
                            className={cn(
                              "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground cursor-pointer group mb-1 text-sm relative",
                              location.pathname.startsWith(`/${model.name}`) &&
                                "bg-accent text-accent-foreground font-medium",
                              sidebarCompact && "justify-center px-2"
                            )}
                            title={
                              sidebarCompact
                                ? model.verbose_name_plural
                                : undefined
                            }
                            aria-label={
                              sidebarCompact
                                ? model.verbose_name_plural
                                : undefined
                            }
                          >
                            {sidebarCompact ? (
                              <div className="h-7 w-7 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-xs font-semibold text-muted-foreground group-hover:text-foreground">
                                {model.verbose_name_plural
                                  .split(" ")
                                  .map((part: string) => part[0])
                                  .join("")
                                  .slice(0, 2)
                                  .toUpperCase()}
                              </div>
                            ) : (
                              <>
                                <Database className="h-4 w-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                                <span className="text-sm flex-1 truncate">
                                  {model.verbose_name_plural}
                                </span>
                              </>
                            )}
                            {(canAdd || !sidebarCompact) && (
                              <div
                                className={cn(
                                  "flex items-center gap-2",
                                  sidebarCompact && "absolute right-2"
                                )}
                              >
                                {canAdd && (
                                  <button
                                    type="button"
                                    aria-label={`Create ${model.verbose_name_plural}`}
                                    onClick={(event) => {
                                      event.stopPropagation();
                                      navigate({
                                        to: "/$model/create",
                                        params: { model: model.name },
                                      });
                                    }}
                                    className="opacity-0 group-hover:opacity-100 transition text-muted-foreground hover:text-foreground h-6 w-6 rounded-md border border-border/60 flex items-center justify-center"
                                  >
                                    <Plus className="h-3 w-3" />
                                  </button>
                                )}
                                {!sidebarCompact && (
                                  <button
                                    type="button"
                                    aria-label={
                                      pinnedModels.includes(model.name)
                                        ? `Unpin ${model.verbose_name_plural}`
                                        : `Pin ${model.verbose_name_plural}`
                                    }
                                    onClick={(event) => {
                                      event.stopPropagation();
                                      togglePinnedModel(model.name);
                                    }}
                                    className={cn(
                                      "opacity-0 group-hover:opacity-100 transition text-muted-foreground hover:text-primary",
                                      pinnedModels.includes(model.name) &&
                                        "opacity-100 text-yellow-500"
                                    )}
                                  >
                                    <Star className="h-4 w-4" />
                                  </button>
                                )}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
                  </div>
                ))}
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
              {pluginSections.map((section) => {
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
                        "flex items-center w-full gap-2 px-3 py-2 rounded-md text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground hover:bg-accent/40 transition-all",
                        sidebarCompact && "justify-center px-2"
                      )}
                      title={sidebarCompact ? section.label : undefined}
                      aria-expanded={isExpanded}
                      aria-controls={`${section.id}-panel`}
                    >
                      <Package className="h-3.5 w-3.5" />
                      {!sidebarCompact && <span>{section.label}</span>}
                      {!sidebarCompact && (
                        <ChevronDown
                          className={cn(
                            "ml-auto h-3 w-3 transition-transform",
                            isExpanded && "rotate-180"
                          )}
                        />
                      )}
                    </button>
                    {isExpanded && (
                      <div
                        id={`${section.id}-panel`}
                        className="space-y-1"
                        role="region"
                        aria-label={section.label}
                      >
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
        </div>

        <div className="absolute bottom-0 left-0 right-0 p-4 bg-card/50 backdrop-blur-sm border-t border-border">
          <Button
            variant="ghost"
            className={cn(
              "w-full text-muted-foreground hover:text-destructive hover:bg-destructive/10",
              sidebarCompact ? "justify-center px-2" : "justify-start"
            )}
            onClick={handleLogout}
            title={sidebarCompact ? "Logout" : undefined}
          >
            <LogOut className={cn("h-4 w-4", !sidebarCompact && "mr-3")} />
            {!sidebarCompact && (
              <span className="text-sm font-medium">Logout</span>
            )}
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-h-screen overflow-hidden bg-muted/20">
        <header className="h-16 border-b border-border/50 flex items-center px-6 bg-card/80 backdrop-blur-md sticky top-0 z-40 shrink-0 gap-4">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden text-muted-foreground"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            aria-label={sidebarOpen ? "Close menu" : "Open menu"}
          >
            {sidebarOpen ? (
              <PanelLeftOpen className="h-5 w-5" />
            ) : (
              <Menu className="h-5 w-5" />
            )}
          </Button>

          <GlobalSearch models={models} />
          {visibleQuickActions.length > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground">
                Quick Actions
              </span>
              <div className="flex items-center gap-2">
                {visibleQuickActions.map((action) => (
                  <Button
                    key={action.label}
                    variant="outline"
                    size="sm"
                    className="h-8 px-3 text-[10px] font-bold uppercase tracking-widest"
                    onClick={action.onClick}
                    disabled={action.disabled}
                    aria-label={action.ariaLabel || action.label}
                  >
                    {action.icon && (
                      <span className="mr-2 flex items-center">
                        {action.icon}
                      </span>
                    )}
                    {action.label}
                  </Button>
                ))}
              </div>
            </div>
          )}

          {showPluginHeader && (
            <div className="hidden md:flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              <span>Plugin</span>
              <ChevronRight className="h-3 w-3" />
              <span className="text-foreground normal-case text-sm font-medium">
                {pluginLabel}
              </span>
              {entryLabel && (
                <>
                  <ChevronRight className="h-3 w-3" />
                  <span className="text-muted-foreground normal-case text-sm">
                    {entryLabel}
                  </span>
                </>
              )}
            </div>
          )}

          <div className="ml-auto flex items-center gap-2">
            <ThemeCustomizer />
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full relative text-muted-foreground hover:text-foreground"
            >
              <Bell className="h-4 w-4" />
              <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-destructive rounded-full border-2 border-card" />
            </Button>
            <div className="h-6 w-[1px] bg-border mx-2" />
            <div className="flex items-center gap-3 pl-2">
              <div className="text-right hidden sm:block">
                <p className="text-sm font-semibold leading-none text-foreground">
                  Admin User
                </p>
                <p className="text-[10px] text-muted-foreground mt-1 uppercase font-bold tracking-wider">
                  Super Admin
                </p>
              </div>
              <div className="w-8 h-8 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-2 ring-background">
                AU
              </div>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <div className="flex-1 p-6 lg:p-8 overflow-auto">
          <div className="mx-auto flex w-full max-w-7xl flex-col gap-8">
            {showPluginHeader && (
              <div className="rounded-lg border border-border bg-card/80 px-4 py-3 shadow-sm">
                <div className="flex flex-wrap items-center gap-3">
                  {showIconGroup && (
                    <div className="flex items-center -space-x-2">
                      {renderIconBadge(
                        activePluginInfo.plugin?.icon,
                        pluginLabel
                      )}
                      {activePluginInfo.entry?.icon &&
                        renderIconBadge(
                          activePluginInfo.entry?.icon,
                          entryLabel
                        )}
                    </div>
                  )}
                  <div className="min-w-0">
                    <p className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                      Plugin
                    </p>
                    <p className="text-lg font-semibold text-foreground truncate">
                      {pluginLabel}
                    </p>
                    {entryLabel && (
                      <p className="text-sm text-muted-foreground truncate">
                        {entryLabel}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}
            {children}
          </div>
        </div>
      </main>
    </div>
  );
}
