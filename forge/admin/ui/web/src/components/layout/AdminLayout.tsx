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
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { cn } from "../../lib/utils";

import { GlobalSearch } from "./GlobalSearch";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { data: modelsData } = useModels();
  const { data: configData } = useConfig();
  const navigate = useNavigate();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [sidebarCompact, setSidebarCompact] = useState(true);
  const [expandedSections, setExpandedSections] = useState<
    Record<string, boolean>
  >({});

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
  };

  const isMenuEntryActive = (entry: any): boolean => {
    if (entry.path && entry.path === location.pathname) return true;
    return entry.children?.some(isMenuEntryActive) ?? false;
  };

  // Recursive Sidebar Item
  const SidebarItem = ({
    item,
    depth = 0,
    compact,
  }: {
    item: any;
    depth?: number;
    compact?: boolean;
  }) => {
    const hasChildren = item.children && item.children.length > 0;
    const [expanded, setExpanded] = useState(false);

    // Auto-expand if active child
    useEffect(() => {
      const isChildActive = (it: any): boolean => {
        if (it.path === location.pathname) return true;
        return it.children?.some(isChildActive) ?? false;
      };
      if (hasChildren && isChildActive(item)) {
        setExpanded(true);
      }
    }, [location.pathname, item, hasChildren]);

    const handleClick = (e: React.MouseEvent) => {
      e.stopPropagation();
      if (hasChildren) {
        setExpanded(!expanded);
      } else {
        const path = item.path.startsWith("/admin/")
          ? item.path.substring(6)
          : item.path;
        navigate({ to: path });
      }
    };

    return (
      <div>
        <div
          onClick={handleClick}
          className={cn(
            "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer text-sm group select-none mb-1",
            location.pathname === item.path &&
              "bg-sidebar-accent text-sidebar-accent-foreground font-medium",
            depth > 0 && "text-muted-foreground"
          )}
          style={{
            paddingLeft: depth === 0 ? "0.75rem" : `${depth * 1 + 0.75}rem`,
          }}
          title={compact ? item.label : undefined}
          aria-label={compact ? item.label : undefined}
        >
          {depth === 0 && (
            <Package className="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0" />
          )}
          {compact && depth > 0 && (
            <div className="h-6 w-6 rounded-md bg-muted/70 border border-border/60 flex items-center justify-center text-[10px] font-semibold text-muted-foreground group-hover:text-foreground">
              {item.label?.charAt(0).toUpperCase()}
            </div>
          )}
          {!compact && <span className="flex-1 truncate">{item.label}</span>}
          {hasChildren && (
            <ChevronDown
              className={cn(
                "h-3 w-3 transition-transform shrink-0 text-muted-foreground",
                expanded && "rotate-180"
              )}
            />
          )}
        </div>
        {hasChildren && expanded && (
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
  const modelGroups = useMemo(() => {
    const groups = new Map<
      string,
      { id: string; label: string; models: any[] }
    >();
    const formatLabel = (value: string) =>
      value
        .replace(/[-_.]/g, " ")
        .replace(/\b\w/g, (char) => char.toUpperCase());

    models.forEach((model: any) => {
      let groupKey = "";
      if (model.name.includes(".")) {
        groupKey = model.name.split(".")[0];
      } else if (model.name.includes("_")) {
        groupKey = model.name.split("_")[0];
      } else {
        groupKey = model.name.charAt(0).toUpperCase();
      }
      const label = formatLabel(groupKey);
      if (!groups.has(groupKey)) {
        groups.set(groupKey, {
          id: `models-${groupKey.toLowerCase()}`,
          label,
          models: [],
        });
      }
      groups.get(groupKey)?.models.push(model);
    });

    return Array.from(groups.values()).sort((a, b) =>
      a.label.localeCompare(b.label)
    );
  }, [models]);

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
    let groupKey = "";
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

  useEffect(() => {
    setExpandedSections((prev) => ({
      ...prev,
      ...(activeModelSectionId ? { [activeModelSectionId]: true } : {}),
      ...(activePluginSectionId ? { [activePluginSectionId]: true } : {}),
    }));
  }, [activeModelSectionId, activePluginSectionId]);

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 w-64 bg-card/95 backdrop-blur-sm border-r border-border transition-all duration-300 ease-in-out lg:relative lg:translate-x-0",
          sidebarCompact ? "lg:w-20" : "lg:w-64",
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
            aria-label={
              sidebarCompact ? "Expand sidebar" : "Collapse sidebar"
            }
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
                  "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
              )}
            >
              <LayoutDashboard className="h-4 w-4" />
              {!sidebarCompact && (
                <span className="font-medium text-sm">Dashboard</span>
              )}
            </Link>
          </div>

          {/* Grouped Sections */}
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
                >
                  <Database className="h-3.5 w-3.5" />
                  {!sidebarCompact && <span>{group.label}</span>}
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
                  <div className="space-y-1">
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
                          "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground cursor-pointer group mb-1 text-sm",
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
                            <span className="text-sm">
                              {model.verbose_name_plural}
                            </span>
                          </>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
          })}

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

        <div className="absolute bottom-0 left-0 right-0 p-4 bg-card/50 backdrop-blur-sm border-t border-border">
          <Button
            variant="ghost"
            className="w-full justify-start text-muted-foreground hover:text-destructive hover:bg-destructive/10"
            onClick={handleLogout}
          >
            <LogOut className="mr-3 h-4 w-4" />
            <span className="text-sm font-medium">Logout</span>
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
          >
            <Menu className="h-5 w-5" />
          </Button>

          <GlobalSearch />

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
        <div className="flex-1 p-8 overflow-auto">
          <div className="mx-auto max-w-7xl animate-in fade-in slide-in-from-bottom-4 duration-500">
            {children}
          </div>
        </div>
      </main>
    </div>
  );
}
