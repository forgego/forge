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
import { Input } from "../components/ui/input";
import { PageHeader } from "../components/ui/page-header";
import { StatTile } from "../components/widgets/stat-tile";

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
          <p className="text-ui text-muted-foreground animate-pulse">
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
          <h2 className="text-display font-bold text-destructive">
            Configuration Error
          </h2>
          <p className="text-ui text-muted-foreground max-w-sm">
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
      <div className="space-y-canvas pb-12">
        <PageHeader
          eyebrow="Overview"
          title="Dashboard"
          description="Registered models, record volume, and installed extensions."
          actions={
            <>
              <Button
                variant="outline"
                size="sm"
                data-testid="quick-find-button"
                className="text-meta gap-1.5 h-9"
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
                className="text-meta gap-1.5 h-9"
              >
                <Link to="/style-guide">
                  <Sparkles className="h-3.5 w-3.5" />
                  Design System
                </Link>
              </Button>
            </>
          }
        />

        <div className="grid grid-cols-1 gap-canvas sm:grid-cols-2 lg:grid-cols-4">
          <StatTile
            className="lg:col-span-2 cursor-pointer"
            label="Registered Models"
            value={models.length}
            description="Schema aggregates managed"
            icon={Database}
            role="link"
            tabIndex={0}
            aria-label={`Browse ${models.length} registered models`}
            onClick={() => navigate({ to: "/registry" })}
            onKeyDown={(e: React.KeyboardEvent) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                navigate({ to: "/registry" });
              }
            }}
          />
          <StatTile
            label="Total Records"
            value={totalRecords.toLocaleString()}
            description="Persistent database rows"
            icon={Layers}
          />
          <StatTile
            label="Active Plugins"
            value={plugins.length}
            description="Modular extensions loaded"
            icon={Package}
          />
          <StatTile
            className="sm:col-span-2 lg:col-span-4"
            label="Audit Logging"
            value="Active"
            description="Thread-safe memory buffer and storage"
            icon={ShieldCheck}
          />
        </div>

        {/* Model Catalog and Live Activity Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-canvas">
          {/* Models Showcase (2 cols) */}
          <div className="lg:col-span-2 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <p className="text-micro text-muted-foreground">
                  Catalog
                </p>
                <h2 className="text-display font-bold tracking-tight text-foreground">
                  Application Models
                </h2>
                <p className="text-meta text-muted-foreground">
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
                  className="h-8 pl-8 pr-8 text-meta bg-background/50 border-border-subtle"
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
              <p className="text-meta text-muted-foreground" aria-live="polite">
                {filteredModels.length} of {models.length} models shown
              </p>
            )}

            {filteredModels.length === 0 ? (
              <div className="rounded-lg border border-dashed border-border p-8 text-center bg-card">
                <Database className="mx-auto h-8 w-8 text-muted-foreground opacity-50" />
                <p className="text-ui font-semibold mt-2">No models found</p>
                <p className="text-meta text-muted-foreground mt-1">
                  No models matched "{modelSearch}".
                </p>
              </div>
            ) : (
              <Card className="overflow-hidden">
                <ul className="divide-y divide-border-subtle">
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
                          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-border-subtle bg-muted/50 text-muted-foreground group-hover:text-primary group-hover:border-primary/30 group-hover:bg-primary/[0.07] transition-colors"
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
                              className="font-semibold text-ui truncate hover:text-primary transition-colors rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              {m.verbose_name_plural}
                            </button>
                            <span className="shrink-0 text-micro tabular-nums text-muted-foreground">
                              {m.count} {m.count === 1 ? "row" : "rows"}
                            </span>
                          </div>
                          <p className="text-meta text-muted-foreground truncate">
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
                            className="h-8 text-meta font-medium text-muted-foreground hover:text-primary px-2.5"
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
                              className="h-8 text-meta gap-1 px-2.5"
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
            <Card className="border-border-subtle">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Activity className="h-4 w-4 text-primary" />
                    <CardTitle className="text-base">System Activity</CardTitle>
                  </div>
                </div>
                <CardDescription className="text-meta">
                  Latest audit events across modules.
                </CardDescription>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="rounded-lg border border-dashed border-border-subtle p-6 text-center">
                  <Activity className="mx-auto h-6 w-6 text-muted-foreground opacity-50" />
                  <p className="mt-2 text-ui font-medium">No activity feed yet</p>
                  <p className="mt-1 text-meta text-muted-foreground">
                    Per-record audit history is available on each record's edit
                    page. A global feed is on the roadmap.
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Quick Tips & Platform Card */}
            <Card className="border-border-subtle">
              <CardHeader className="pb-2">
                <CardTitle className="text-ui font-semibold flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-success" />
                  Forge Framework Architecture
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-meta text-muted-foreground">
                <p>
                  Full-stack Go administrative framework with schema-driven validation, server-driven UI (SDUI), and immutable audit history.
                </p>
                <div className="pt-2 flex flex-col gap-1.5 font-mono text-micro">
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Codegen REST APIs</span>
                    <span className="text-success">Available</span>
                  </div>
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Migration Recovery</span>
                    <span className="text-success">Supported</span>
                  </div>
                  <div className="flex items-center justify-between text-foreground/80">
                    <span>Audit Manager</span>
                    <span className="text-success">Active</span>
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
