import React, { useMemo, useState, useEffect, useCallback } from "react";
import { useReducedMotion } from "framer-motion";
import { useNavigate, useLocation } from "@tanstack/react-router";
import { useModels, useConfig, useLogout } from "../../api/hooks/adminHooks";
import { Breadcrumbs } from "./Breadcrumbs";
import { useShortcutHelp, ShortcutHelpDialog } from "../../hooks/useKeyboardShortcuts";
import { isMenuEntryActive, findActiveMenuEntry } from "./nav-utils";
import { Sidebar } from "./Sidebar";
import { TopBar } from "./TopBar";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
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
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({});
  const [modelFilter, setModelFilter] = useState("");
  const { showHelp, setShowHelp } = useShortcutHelp();
  const prefersReducedMotion = useReducedMotion();

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token && location.pathname !== "/login") {
      navigate({ to: "/login" });
    }
  }, [navigate, location.pathname]);

  // Close the mobile drawer on navigation and on Escape.
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- sync sidebar state with pathname changes
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

  const go = useCallback((to: string, params?: Record<string, string>) => {
    setSidebarOpen(false);
    navigate(params ? ({ to, params } as any) : ({ to } as any));
  }, [navigate]);

  const handleLogout = () => {
    logoutMutation.mutate(undefined, { onSettled: () => navigate({ to: "/login" }) });
  };

  const models = modelsData?.models || [];
  const plugins = configData?.plugins || [];
  const sessionUser = (configData as any)?.user as { name?: string; username?: string; role?: string } | undefined;
  const userName = sessionUser?.name || sessionUser?.username || "Admin";
  const userRole = sessionUser?.role || "Super Admin";
  const userInitials = userName.split(/[\s_.-]+/).map((part) => part.charAt(0)).join("").slice(0, 2).toUpperCase();

  useEffect(() => {
    localStorage.setItem("forge.admin.pinnedModels", JSON.stringify(pinnedModels));
  }, [pinnedModels]);

  const togglePinnedModel = (modelName: string) => {
    setPinnedModels((prev) => (prev.includes(modelName) ? prev.filter((name) => name !== modelName) : [...prev, modelName]));
  };

  const formatGroupLabel = (value: string) => value.replace(/[_-]+/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());

  const groupByModel = useMemo(() => {
    const configGroups = configData?.model_groups ?? configData?.modelGroups ?? configData?.model_groups_by_app ?? null;
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
      .filter((model: any) => query === "" || model.verbose_name_plural?.toLowerCase().includes(query) || model.name?.toLowerCase().includes(query))
      .forEach((model: any) => {
        const configLabel = modelGroupMap.get(model.name);
        const derivedGroup = (() => {
          if (model.name.includes(".")) {
            return model.name.split(".")[0];
          }
          if (model.name.includes("__")) {
            return model.name.split("__")[0];
          }
          return ""; // no group
        })();
        const label = formatGroupLabel(configLabel || derivedGroup);
        const existing = grouped.get(label) || [];
        existing.push(model);
        grouped.set(label, existing);
      });

    return Array.from(grouped.entries())
      .map(([label, groupModels]) => ({
        label,
        models: groupModels.sort((a: any, b: any) => a.verbose_name_plural.localeCompare(b.verbose_name_plural)),
      }))
      .sort((a, b) => {
        if (a.label === "" && b.label !== "") return -1;
        if (a.label !== "" && b.label === "") return 1;
        return a.label.localeCompare(b.label);
      });
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
      plugin.menuEntries?.some((entry: any) => isMenuEntryActive(entry, location.pathname))
    );
    return activePlugin ? `plugin-${activePlugin.name}` : null;
  }, [location.pathname, plugins]);

  // eslint-disable-next-line react-hooks/preserve-manual-memoization -- module-level functions are stable, React Compiler cannot track them
  const activePluginInfo = useMemo(() => {
    for (const plugin of plugins) {
      const entries = plugin.menuEntries ?? [];
      const activeEntry = findActiveMenuEntry(entries, location.pathname);
      if (activeEntry) {
        return { plugin, entry: activeEntry };
      }
    }
    return { plugin: null, entry: null };
  }, [plugins, location.pathname]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- sync expanded sections with active section changes
    setExpandedSections((prev) => ({
      ...prev,
      ...(activeModelSectionId ? { [activeModelSectionId]: true } : {}),
      ...(activePluginSectionId ? { [activePluginSectionId]: true } : {}),
    }));
  }, [activeModelSectionId, activePluginSectionId]);

  const showPluginHeader = Boolean(activePluginInfo.plugin);
  const pluginLabel = activePluginInfo.plugin?.label || activePluginInfo.plugin?.name;
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
      <Sidebar
        models={models} groupByModel={groupByModel} pluginSections={pluginSections}
        pinnedModels={pinnedModels} togglePinnedModel={togglePinnedModel}
        modelFilter={modelFilter} setModelFilter={setModelFilter}
        expandedSections={expandedSections} setExpandedSections={setExpandedSections}
        sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen}
        sidebarCompact={sidebarCompact} setSidebarCompact={setSidebarCompact}
        pathname={location.pathname} go={go} onLogout={handleLogout}
        logoutPending={logoutMutation.isPending} prefersReducedMotion={prefersReducedMotion}
      />
      {/* Main Content */}
      <main className="flex-1 flex flex-col min-h-screen min-w-0 bg-muted/20">
        <TopBar
          models={models} userName={userName} userRole={userRole} userInitials={userInitials}
          onOpenMobileNav={() => setSidebarOpen(true)} onShowHelp={() => setShowHelp(true)}
          showPluginHeader={showPluginHeader} pluginLabel={pluginLabel} entryLabel={entryLabel}
        />
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
