import { useState, useMemo } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import {
  Database,
  Plus,
  Search,
  ArrowRight,
  ShieldCheck,
  Layers,
  Activity,
  Command,
  Loader2,
  Sparkles,
  CheckCircle2,
  Package,
  X,
} from "lucide-react";
import AdminLayout from "../components/layout/AdminLayout";
import { ModelIcon } from "../components/ModelIcon";
import { useConfig, useMetadata } from "../api/hooks/adminHooks";
import { SDUIRoot } from "../components/sdui/DynamicRenderer";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Badge } from "../components/ui/badge";
import { Input } from "../components/ui/input";

export default function DashboardPage() {
  const { data: config, isLoading: configLoading, error: configError } = useConfig();
  const { data: metadata, isLoading: metaLoading } = useMetadata();
  const navigate = useNavigate();

  const [modelSearch, setModelSearch] = useState("");

  const models = metadata?.models || [];
  const plugins = metadata?.plugins || [];

  const totalRecords = useMemo(() => {
    return models.reduce((acc, m) => acc + (m.count || 0), 0);
  }, [models]);

  const filteredModels = useMemo(() => {
    if (!modelSearch.trim()) return models;
    const q = modelSearch.toLowerCase();
    return models.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.verbose_name.toLowerCase().includes(q) ||
        m.verbose_name_plural.toLowerCase().includes(q)
    );
  }, [models, modelSearch]);

  if (configLoading || metaLoading) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground animate-pulse">
            Loading administrative environment...
          </p>
        </div>
      </AdminLayout>
    );
  }

  if (configError) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-3 text-center">
          <div className="rounded-full bg-destructive/10 p-3 text-destructive">
            <Database className="h-6 w-6" />
          </div>
          <h2 className="text-2xl font-bold text-destructive">
            Configuration Error
          </h2>
          <p className="text-sm text-muted-foreground max-w-sm">
            Failed to load admin system configuration. Please check backend service status.
          </p>
        </div>
      </AdminLayout>
    );
  }

  // If a custom SDUI dashboard layout is provided by the backend, render it!
  const dashboardConfig = config?.dashboard;
  if (dashboardConfig && dashboardConfig.Layout) {
    return (
      <AdminLayout>
        <SDUIRoot component={dashboardConfig.Layout} />
      </AdminLayout>
    );
  }

  // Default Executive Dashboard
  return (
    <AdminLayout>
      <div className="space-y-8 pb-12">
        {/* Welcome & System Status Banner */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border/40 pb-6">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className="border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 gap-1.5 text-xs font-medium"
              >
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </span>
                System Operational
              </Badge>
              <Badge variant="secondary" className="text-xs font-mono">
                {config?.version ? `Forge v${config.version}` : "Forge"}
              </Badge>
            </div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              Admin Console
            </h1>
            <p className="text-sm text-muted-foreground">
              Unified control panel for schema entities, audit trails, and administrative workflows.
            </p>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            <Button
              variant="outline"
              size="sm"
              data-testid="quick-find-button"
              className="text-xs gap-1.5 h-9"
              onClick={() => {
                (window as any).__forgeOpenSearch?.();
              }}
            >
              <Command className="h-3.5 w-3.5" />
              Quick Find (⌘K)
            </Button>
            <Button
              variant="default"
              size="sm"
              asChild
              className="text-xs gap-1.5 h-9"
            >
              <Link to="/style-guide">
                <Sparkles className="h-3.5 w-3.5" />
                Design System
              </Link>
            </Button>
          </div>
        </div>

        {/* Executive Stats Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card
            className="border-border/50 shadow-sm hover:border-primary/30 transition-colors cursor-pointer"
            role="link"
            tabIndex={0}
            aria-label={`Browse ${models.length} registered models`}
            onClick={() => navigate({ to: "/registry" })}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                navigate({ to: "/registry" });
              }
            }}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Registered Models
              </CardTitle>
              <Database className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">{models.length}</div>
              <p className="text-xs text-muted-foreground mt-1">
                Schema aggregates managed
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/50 shadow-sm hover:border-primary/30 transition-colors">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Total Records
              </CardTitle>
              <Layers className="h-4 w-4 text-emerald-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">
                {totalRecords.toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Persistent database rows
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/50 shadow-sm hover:border-primary/30 transition-colors">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Audit Logging
              </CardTitle>
              <ShieldCheck className="h-4 w-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight text-blue-600 dark:text-blue-400">
                Active
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Thread-safe memory buffer & storage
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/50 shadow-sm hover:border-primary/30 transition-colors">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Active Plugins
              </CardTitle>
              <Package className="h-4 w-4 text-purple-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">
                {plugins.length}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Modular extensions loaded
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Model Catalog and Live Activity Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Models Showcase (2 cols) */}
          <div className="lg:col-span-2 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h2 className="text-xl font-bold tracking-tight text-foreground">
                  Application Models
                </h2>
                <p className="text-xs text-muted-foreground">
                  Direct entry points for CRUD operations, filtering, and bulk actions.
                </p>
              </div>

              <div className="relative w-full sm:w-64">
                <label htmlFor="dashboard-model-filter" className="sr-only">
                  Filter application models
                </label>
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
                <Input
                  id="dashboard-model-filter"
                  value={modelSearch}
                  onChange={(e) => setModelSearch(e.target.value)}
                  placeholder="Filter models..."
                  className="h-8 pl-8 pr-8 text-xs bg-background/50 border-border/60"
                />
                {modelSearch && (
                  <button
                    type="button"
                    aria-label="Clear model filter"
                    onClick={() => setModelSearch("")}
                    className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>
            {modelSearch.trim() && (
              <p className="text-xs text-muted-foreground" aria-live="polite">
                {filteredModels.length} of {models.length} models shown
              </p>
            )}

            {filteredModels.length === 0 ? (
              <div className="rounded-xl border border-dashed border-border p-8 text-center bg-card">
                <Database className="mx-auto h-8 w-8 text-muted-foreground opacity-50" />
                <p className="text-sm font-semibold mt-2">No models found</p>
                <p className="text-xs text-muted-foreground mt-1">
                  No models matched "{modelSearch}".
                </p>
              </div>
            ) : (
              <Card className="overflow-hidden">
                <ul className="divide-y divide-border/60">
                  {filteredModels.map((m) => {
                    const caps = [
                      m.permissions.view && "Read",
                      m.permissions.add && "Create",
                      m.permissions.change && "Update",
                      m.permissions.delete && "Delete",
                    ].filter(Boolean);
                    return (
                      <li
                        key={m.name}
                        className="group flex items-center gap-3 px-4 py-3 hover:bg-muted/40 transition-colors"
                      >
                        <span
                          aria-hidden
                          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-border/60 bg-muted/50 text-muted-foreground group-hover:text-primary group-hover:border-primary/30 group-hover:bg-primary/[0.07] transition-colors"
                        >
                          <ModelIcon name={m.icon} className="h-4 w-4" />
                        </span>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-baseline gap-2">
                            <button
                              type="button"
                              onClick={() =>
                                navigate({
                                  to: "/$model",
                                  params: { model: m.name },
                                })
                              }
                              className="font-semibold text-[14px] truncate hover:text-primary transition-colors rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              {m.verbose_name_plural}
                            </button>
                            <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                              {m.count} {m.count === 1 ? "row" : "rows"}
                            </span>
                          </div>
                          <p className="text-xs text-muted-foreground truncate">
                            <span className="font-mono">{m.name}</span>
                            {caps.length > 0 && (
                              <span> · {caps.join(" · ")}</span>
                            )}
                          </p>
                        </div>
                        <div className="flex items-center gap-1 shrink-0 opacity-100 focus-within:opacity-100 lg:opacity-0 lg:group-hover:opacity-100 lg:group-focus-within:opacity-100 transition-opacity">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 text-xs font-medium text-muted-foreground hover:text-primary px-2.5"
                            onClick={() =>
                              navigate({
                                to: "/$model",
                                params: { model: m.name },
                              })
                            }
                          >
                            Browse
                            <ArrowRight className="h-3.5 w-3.5 ml-1" />
                          </Button>
                          {m.permissions.add && (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 text-xs gap-1 px-2.5"
                              onClick={() =>
                                navigate({
                                  to: "/$model/create",
                                  params: { model: m.name },
                                })
                              }
                            >
                              <Plus className="h-3.5 w-3.5" />
                              New
                            </Button>
                          )}
                        </div>
                      </li>
                    );
                  })}
                </ul>
              </Card>
            )}
          </div>

          {/* Activity & Quick Shortcuts (1 col) */}
          <div className="space-y-6">
            {/* Recent Activity Card */}
            <Card className="border-border/50 shadow-sm">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Activity className="h-4 w-4 text-primary" />
                    <CardTitle className="text-base">System Activity</CardTitle>
                  </div>
                </div>
                <CardDescription className="text-xs">
                  Latest audit events across modules.
                </CardDescription>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="rounded-lg border border-dashed border-border/70 p-6 text-center">
                  <Activity className="mx-auto h-6 w-6 text-muted-foreground opacity-50" />
                  <p className="mt-2 text-sm font-medium">No activity feed yet</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Per-record audit history is available on each record's edit
                    page. A global feed is on the roadmap.
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Quick Tips & Platform Card */}
            <Card className="border-border/50 bg-gradient-to-br from-card to-muted/20 shadow-sm">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-semibold flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                  Forge Framework Architecture
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-xs text-muted-foreground">
                <p>
                  Full-stack Go administrative framework with schema-driven validation, server-driven UI (SDUI), and immutable audit history.
                </p>
                <div className="pt-2 flex flex-col gap-1.5 font-mono text-[11px]">
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Codegen REST APIs</span>
                    <span className="text-emerald-600 dark:text-emerald-400">Available</span>
                  </div>
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Migration Recovery</span>
                    <span className="text-emerald-600 dark:text-emerald-400">Supported</span>
                  </div>
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Audit Manager</span>
                    <span className="text-emerald-600 dark:text-emerald-400">Active</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
}
