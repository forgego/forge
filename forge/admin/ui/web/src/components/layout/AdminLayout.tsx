import React from "react";
import { Link, useLocation, useNavigate } from "@tanstack/react-router";
import { LayoutDashboard, LogOut, Menu } from "lucide-react";

import { useConfig, useModels } from "../../api/hooks/adminHooks";
import { cn } from "../../lib/utils";
import { Button } from "../ui/button";
import { GlobalSearch } from "./GlobalSearch";

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

  const models = modelsData?.models ?? [];
  const plugins = configData?.plugins ?? [];

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
  };

  const isActive = (path: string) => location.pathname === path;

  const sortedModels = [...models].sort((a: any, b: any) =>
    a.verbose_name_plural.localeCompare(b.verbose_name_plural)
  );

  return (
    <div className="min-h-screen bg-background flex">
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-64 bg-card border-r border-border transition-transform lg:relative lg:translate-x-0",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="h-16 flex items-center justify-between px-4 border-b border-border">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-md bg-primary text-primary-foreground flex items-center justify-center font-bold">
              F
            </div>
            <span className="text-lg font-semibold">Forge Admin</span>
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
        </div>

        <div className="p-4 space-y-4 overflow-y-auto max-h-[calc(100vh-64px)]">
          <GlobalSearch models={models} />

          <div className="space-y-1">
            <Link
              to="/"
              className={cn(
                "flex items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                isActive("/") && "bg-muted font-medium"
              )}
            >
              <LayoutDashboard className="h-4 w-4" />
              Dashboard
            </Link>
          </div>

          <div>
            <p className="px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Models
            </p>
            <div className="mt-2 space-y-1">
              {sortedModels.map((model: any) => (
                <Link
                  key={model.name}
                  to="/$model"
                  params={{ model: model.name }}
                  className={cn(
                    "block rounded-md px-3 py-2 text-sm hover:bg-muted",
                    isActive(`/${model.name}`) && "bg-muted font-medium"
                  )}
                >
                  {model.verbose_name_plural}
                </Link>
              ))}
            </div>
          </div>

          {plugins.length > 0 && (
            <div>
              <p className="px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Plugins
              </p>
              <div className="mt-2 space-y-1">
                {plugins.flatMap((plugin: any) =>
                  (plugin.menuEntries ?? []).map((entry: any) => (
                    <Link
                      key={`${plugin.id}-${entry.path}`}
                      to={entry.path.replace("/admin", "")}
                      className={cn(
                        "block rounded-md px-3 py-2 text-sm hover:bg-muted",
                        isActive(entry.path.replace("/admin", "")) &&
                          "bg-muted font-medium"
                      )}
                    >
                      {entry.label}
                    </Link>
                  ))
                )}
              </div>
            </div>
          )}
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-h-screen">
        <header className="h-16 border-b border-border flex items-center justify-between px-4">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open sidebar"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <div className="flex items-center gap-2">
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
