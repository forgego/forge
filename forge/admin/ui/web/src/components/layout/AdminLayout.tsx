import React from "react";
import { Link, useLocation, useNavigate } from "@tanstack/react-router";
import {
  Bell,
  ChevronLeft,
  ChevronRight,
  LayoutDashboard,
  LogOut,
  Menu,
  Star,
} from "lucide-react";

import { useConfig, useModels } from "../../api/hooks/adminHooks";
import { cn } from "../../lib/utils";
import { Button } from "../ui/button";
import { GlobalSearch } from "./GlobalSearch";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";

const storageKey = "forge.admin.pinnedModels";

type ModelGroup = {
  id: string;
  label: string;
  models: any[];
};

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { data: modelsData } = useModels();
  const { data: configData } = useConfig();
  const location = useLocation();
  const navigate = useNavigate();

  const [sidebarOpen, setSidebarOpen] = React.useState(true);
  const [sidebarCompact, setSidebarCompact] = React.useState(false);
  const [pinnedModels, setPinnedModels] = React.useState<string[]>(() => {
    if (typeof window === "undefined") return [];
    const stored = localStorage.getItem(storageKey);
    if (!stored) return [];
    try {
      const parsed = JSON.parse(stored);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  });

  const models = modelsData?.models ?? [];
  const plugins = configData?.plugins ?? [];
  const currentPath = location.pathname.replace(/^\/admin/, "") || "/";

  React.useEffect(() => {
    localStorage.setItem(storageKey, JSON.stringify(pinnedModels));
  }, [pinnedModels]);

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
  };

  const isActive = (path: string) => currentPath === path;

  const handleTogglePin = (modelName: string) => {
    setPinnedModels((prev) =>
      prev.includes(modelName)
        ? prev.filter((name) => name !== modelName)
        : [...prev, modelName]
    );
  };

  const groupedModels = React.useMemo<ModelGroup[]>(() => {
    const groups = new Map<string, ModelGroup>();

    const formatLabel = (value: string) =>
      value
        .replace(/[_-]+/g, " ")
        .replace(/\b\w/g, (char) => char.toUpperCase());

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

    models.forEach((model: any) => {
      let groupKey = modelGroupMap.get(model.name) ?? "Other";
      if (!modelGroupMap.has(model.name)) {
        if (model.name.includes(".")) {
          groupKey = model.name.split(".")[0];
        } else if (model.name.includes("_")) {
          groupKey = model.name.split("_")[0];
        }
      }

      const id = `models-${groupKey.toLowerCase()}`;
      if (!groups.has(id)) {
        groups.set(id, {
          id,
          label: formatLabel(groupKey),
          models: [],
        });
      }
      groups.get(id)?.models.push(model);
    });

    return Array.from(groups.values())
      .map((group) => ({
        ...group,
        models: group.models.sort((a: any, b: any) =>
          a.verbose_name_plural.localeCompare(b.verbose_name_plural)
        ),
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [configData, models]);

  const pinned = React.useMemo(
    () =>
      pinnedModels
        .map((name) => models.find((model: any) => model.name === name))
        .filter(Boolean),
    [models, pinnedModels]
  );

  const normalizedPluginEntries = React.useMemo(
    () =>
      plugins.flatMap((plugin: any) =>
        (plugin.menuEntries ?? []).map((entry: any) => ({
          ...entry,
          path: entry.path.replace("/admin", ""),
          pluginId: plugin.id,
        }))
      ),
    [plugins]
  );

  return (
    <div className="min-h-screen bg-background flex">
      {sidebarOpen && (
        <button
          type="button"
          className="fixed inset-0 z-30 bg-background/60 backdrop-blur-sm lg:hidden"
          aria-label="Close sidebar"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 bg-card border-r border-border transition-all duration-200 lg:relative lg:translate-x-0",
          sidebarCompact ? "w-20" : "w-64",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div
          className={cn(
            "h-16 flex items-center border-b border-border",
            sidebarCompact ? "px-3" : "px-4"
          )}
        >
          <div className="flex items-center gap-2 flex-1">
            <div className="w-8 h-8 rounded-md bg-primary text-primary-foreground flex items-center justify-center font-bold">
              F
            </div>
            {!sidebarCompact && (
              <span className="text-lg font-semibold">Forge Admin</span>
            )}
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setSidebarOpen(false)}
            aria-label="Close sidebar"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="hidden lg:inline-flex"
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

        <div className="p-4 space-y-4 overflow-y-auto max-h-[calc(100vh-64px)]">
          <GlobalSearch
            models={models}
            compact={sidebarCompact}
            triggerLabel="Search models"
            className={sidebarCompact ? "w-full" : ""}
          />

          <div className="space-y-1">
            <Link
              to="/"
              className={cn(
                "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                isActive("/") && "bg-muted font-medium",
                sidebarCompact && "justify-center"
              )}
            >
              <LayoutDashboard className="h-4 w-4" />
              {!sidebarCompact && "Dashboard"}
            </Link>
          </div>

          {pinned.length > 0 && (
            <div>
              {!sidebarCompact && (
                <p className="px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Pinned
                </p>
              )}
              <div className="mt-2 space-y-1">
                {pinned.map((model: any) => (
                  <Link
                    key={model.name}
                    to="/$model"
                    params={{ model: model.name }}
                    className={cn(
                      "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                      isActive(`/${model.name}`) && "bg-muted font-medium",
                      sidebarCompact && "justify-center"
                    )}
                  >
                    <Star className="h-4 w-4 text-yellow-500" />
                    {!sidebarCompact && model.verbose_name_plural}
                  </Link>
                ))}
              </div>
            </div>
          )}

          <div>
            {!sidebarCompact && (
              <p className="px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Models
              </p>
            )}
            <div className="mt-2 space-y-4">
              {groupedModels.map((group) => (
                <div key={group.id}>
                  {!sidebarCompact && (
                    <p className="px-3 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                      {group.label}
                    </p>
                  )}
                  <div className="mt-1 space-y-1">
                    {group.models.map((model: any) => {
                      const active = isActive(`/${model.name}`);
                      return (
                        <div
                          key={model.name}
                          className={cn(
                            "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                            active && "bg-muted font-medium",
                            sidebarCompact && "justify-center"
                          )}
                        >
                          <Link
                            to="/$model"
                            params={{ model: model.name }}
                            className={cn(
                              "flex-1 truncate",
                              sidebarCompact && "text-center"
                            )}
                          >
                            {sidebarCompact
                              ? model.verbose_name_plural
                                  .charAt(0)
                                  .toUpperCase()
                              : model.verbose_name_plural}
                          </Link>
                          {!sidebarCompact && (
                            <button
                              type="button"
                              onClick={() => handleTogglePin(model.name)}
                              aria-label="Pin model"
                              className="text-muted-foreground hover:text-foreground"
                            >
                              <Star
                                className={cn(
                                  "h-4 w-4",
                                  pinnedModels.includes(model.name)
                                    ? "text-yellow-500"
                                    : "text-muted-foreground"
                                )}
                              />
                            </button>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {normalizedPluginEntries.length > 0 && (
            <div>
              {!sidebarCompact && (
                <p className="px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Plugins
                </p>
              )}
              <div className="mt-2 space-y-1">
                {normalizedPluginEntries.map((entry: any) => (
                  <Link
                    key={`${entry.pluginId}-${entry.path}`}
                    to={entry.path}
                    className={cn(
                      "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                      isActive(entry.path) && "bg-muted font-medium",
                      sidebarCompact && "justify-center"
                    )}
                  >
                    {!sidebarCompact && entry.label}
                    {sidebarCompact && entry.label?.charAt(0).toUpperCase()}
                  </Link>
                ))}
              </div>
            </div>
          )}
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-h-screen">
        <header className="h-16 border-b border-border flex items-center justify-between px-4 gap-4 bg-card/80 backdrop-blur-md sticky top-0 z-40">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open sidebar"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <div className="flex-1 max-w-xl hidden md:block">
            <GlobalSearch models={models} />
          </div>
          <div className="flex items-center gap-2">
            <ThemeCustomizer />
            <Button
              variant="ghost"
              size="icon"
              className="rounded-full relative text-muted-foreground hover:text-foreground"
              aria-label="Notifications"
            >
              <Bell className="h-4 w-4" />
              <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-destructive rounded-full border-2 border-card" />
            </Button>
            <div className="hidden sm:flex items-center gap-3 pl-2">
              <div className="text-right">
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
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="text-muted-foreground"
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </Button>
          </div>
        </header>

        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
