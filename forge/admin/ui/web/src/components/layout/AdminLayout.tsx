import React, { useState, useEffect } from "react";
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

  const handleLogout = () => {
    localStorage.removeItem("admin_token");
    navigate({ to: "/login" });
  };

  // Recursive Sidebar Item
  const SidebarItem = ({ item, depth = 0 }: { item: any; depth?: number }) => {
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
        >
          {depth === 0 && (
            <Package className="h-4 w-4 text-muted-foreground group-hover:text-foreground shrink-0" />
          )}
          <span className="flex-1 truncate">{item.label}</span>
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
              <SidebarItem key={idx} item={child} depth={depth + 1} />
            ))}
          </div>
        )}
      </div>
    );
  };

  const models = modelsData?.models || [];
  const plugins = configData?.plugins || [];

  return (
    <div className="min-h-screen bg-background flex">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 w-64 bg-card/95 backdrop-blur-sm border-r border-border transition-transform duration-300 ease-in-out lg:relative lg:translate-x-0",
          !sidebarOpen && "-translate-x-full lg:hidden"
        )}
      >
        <div className="h-16 flex items-center px-6 border-b border-border/50">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-primary-foreground font-bold shadow-sm">
              F
            </div>
            <span className="text-lg font-bold tracking-tight text-foreground">
              Forge Admin
            </span>
          </div>
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
                  "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
              )}
            >
              <LayoutDashboard className="h-4 w-4" />
              <span className="font-medium text-sm">Dashboard</span>
            </Link>
          </div>

          {/* Models Section */}
          <div className="space-y-1">
            <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
              Content Models
            </h4>
            {models.map((model: any) => (
              <div
                key={model.name}
                data-testid={`nav-${model.name}`}
                onClick={() => {
                  navigate({ to: "/$model", params: { model: model.name } });
                }}
                className={cn(
                  "flex items-center gap-3 px-3 py-2 rounded-md transition-all hover:bg-accent hover:text-accent-foreground cursor-pointer group mb-1",
                  location.pathname.startsWith(`/${model.name}`) &&
                    "bg-accent text-accent-foreground font-medium"
                )}
              >
                <Database className="h-4 w-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                <span className="text-sm">{model.verbose_name_plural}</span>
              </div>
            ))}
          </div>

          {/* Plugins Section */}
          {plugins.length > 0 && (
            <div className="space-y-1">
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2">
                Plugins
              </h4>
              {plugins.map((plugin: any) => (
                <React.Fragment key={plugin.name}>
                  {plugin.menuEntries?.map((entry: any, idx: number) => (
                    <SidebarItem
                      key={`${plugin.name}-menu-${idx}`}
                      item={entry}
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
        <div className="flex-1 p-6 lg:p-8 overflow-auto">
          <div className="mx-auto flex w-full max-w-7xl flex-col gap-8">
            {children}
          </div>
        </div>
      </main>
    </div>
  );
}
