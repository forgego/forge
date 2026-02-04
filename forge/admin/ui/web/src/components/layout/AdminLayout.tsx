import React, { useState, useEffect, useMemo } from "react";
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
  PanelLeftClose,
  PanelLeftOpen,
  Star,
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
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);
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

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
  };

  // Recursive Sidebar Item
  const SidebarItem = ({
    item,
    depth = 0,
    collapsed = false,
  }: {
    item: any;
    depth?: number;
    collapsed?: boolean;
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
        if (!collapsed) {
          setExpanded(!expanded);
        }
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
            depth > 0 && "text-muted-foreground",
            collapsed && "justify-center px-2"
          )}
          style={{
            paddingLeft:
              depth === 0 || collapsed
                ? "0.75rem"
                : `${depth * 1 + 0.75}rem`,
          }}
          title={collapsed ? item.label : undefined}
        >
          {depth === 0 && (
            <Package className="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0" />
          )}
          {!collapsed && <span className="flex-1 truncate">{item.label}</span>}
          {hasChildren && !collapsed && (
            <ChevronDown
              className={cn(
                "h-3 w-3 transition-transform shrink-0 text-muted-foreground",
                expanded && "rotate-180"
              )}
            />
          )}
        </div>
        {hasChildren && expanded && !collapsed && (
          <div className="space-y-1 pt-1">
            {item.children.map((child: any, idx: number) => (
              <SidebarItem
                key={idx}
                item={child}
                depth={depth + 1}
                collapsed={collapsed}
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
  }, [configData, models]);

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 bg-card/95 backdrop-blur-sm border-r border-border transition-all duration-300 ease-in-out lg:relative lg:translate-x-0",
          sidebarCollapsed ? "w-20" : "w-64",
          !sidebarOpen && "-translate-x-full lg:translate-x-0"
        )}
      >
        <div
          className={cn(
            "h-16 flex items-center border-b border-border/50",
            sidebarCollapsed ? "px-4 justify-center" : "px-6"
          )}
        >
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground font-bold shadow-sm">
              F
            </div>
            {!sidebarCollapsed && (
              <span className="text-lg font-bold tracking-tight text-foreground">
                Forge Admin
              </span>
            )}
          </div>
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              "ml-auto text-muted-foreground hover:text-foreground",
              sidebarCollapsed && "hidden"
            )}
            onClick={() => setSidebarCollapsed(true)}
          >
            <PanelLeftClose className="h-4 w-4" />
          </Button>
        </div>
        <div className="p-4 space-y-6 overflow-y-auto max-h-[calc(100vh-140px)] no-scrollbar">
          {/* Main Nav */}
          <div className="space-y-1">
            <Link
              to="/"
              data-testid="nav-dashboard"
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground group mb-1",
                location.pathname === "/" &&
                  "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm",
                sidebarCollapsed && "justify-center px-2"
              )}
              title={sidebarCollapsed ? "Dashboard" : undefined}
            >
              <LayoutDashboard className="h-4 w-4" />
              {!sidebarCollapsed && (
                <span className="font-medium text-sm">Dashboard</span>
              )}
            </Link>
          </div>

          {sidebarCollapsed && (
            <Button
              variant="ghost"
              size="icon"
              className="w-full text-muted-foreground hover:text-foreground"
              onClick={() => setSidebarCollapsed(false)}
            >
              <PanelLeftOpen className="h-4 w-4" />
            </Button>
          )}

          {/* Pinned Models */}
          {pinnedModels.length > 0 && (
            <div className="space-y-1">
              {!sidebarCollapsed && (
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
                      sidebarCollapsed && "justify-center px-2"
                    )}
                    title={sidebarCollapsed ? model.verbose_name_plural : undefined}
                  >
                    <Star className="h-4 w-4 text-yellow-500" />
                    {!sidebarCollapsed && (
                      <span className="text-sm">{model.verbose_name_plural}</span>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {/* Models Section */}
          <div className="space-y-4">
            {!sidebarCollapsed && (
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3">
                Content Models
              </h4>
            )}
            {groupByModel.map((group) => (
              <div key={group.label} className="space-y-1">
                {!sidebarCollapsed && (
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
                      sidebarCollapsed && "justify-center px-2"
                    )}
                    title={
                      sidebarCollapsed ? model.verbose_name_plural : undefined
                    }
                  >
                    <Database className="h-4 w-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                    {!sidebarCollapsed && (
                      <span className="text-sm flex-1 truncate">
                        {model.verbose_name_plural}
                      </span>
                    )}
                    {!sidebarCollapsed && (
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
                ))}
              </div>
            ))}
          </div>

          {/* Plugins Section */}
          {plugins.length > 0 && (
            <div className="space-y-1">
              {!sidebarCollapsed && (
                <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                  Plugins
                </h4>
              )}
              {plugins.map((plugin: any) => (
                <React.Fragment key={plugin.name}>
                  {plugin.menuEntries?.map((entry: any, idx: number) => (
                    <SidebarItem
                      key={`${plugin.name}-menu-${idx}`}
                      item={entry}
                      collapsed={sidebarCollapsed}
                    />
                  ))}
                </React.Fragment>
              ))}
            </div>
          )}
        </div>

        <div className="absolute bottom-0 left-0 right-0 p-4 bg-card/50 backdrop-blur-sm border-t border-border">
          <Button
            variant="ghost"
            className={cn(
              "w-full text-muted-foreground hover:text-destructive hover:bg-destructive/10",
              sidebarCollapsed ? "justify-center px-2" : "justify-start"
            )}
            onClick={handleLogout}
            title={sidebarCollapsed ? "Logout" : undefined}
          >
            <LogOut className={cn("h-4 w-4", !sidebarCollapsed && "mr-3")} />
            {!sidebarCollapsed && (
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
          >
            <Menu className="h-5 w-5" />
          </Button>

          <GlobalSearch models={models} />

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
