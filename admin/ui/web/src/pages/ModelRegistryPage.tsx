import { useState, useMemo } from "react";
import { useModels } from "../api/hooks/adminHooks";
import {
  Card,
} from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Database, Loader2, Plus, Search, ArrowRight } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import AdminLayout from "../components/layout/AdminLayout";
import { ModelIcon } from "../components/ModelIcon";

export default function ModelsListPage() {
  const { data, isLoading, error } = useModels();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");

  const models = data?.models || [];

  const filteredModels = useMemo(() => {
    if (!searchQuery.trim()) return models;
    const q = searchQuery.toLowerCase();
    return models.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.verbose_name.toLowerCase().includes(q) ||
        m.verbose_name_plural.toLowerCase().includes(q)
    );
  }, [models, searchQuery]);

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground animate-pulse">
            Loading model registry...
          </p>
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-3 text-center">
          <h2 className="text-2xl font-bold text-destructive">Error</h2>
          <p className="text-muted-foreground max-w-sm">{error.message}</p>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6 pb-12">
        {/* Header and Search */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              Model Registry
            </h1>
            <p className="text-sm text-muted-foreground">
              Explore and manage registered application data models and schemas.
            </p>
          </div>

          <div className="relative w-full sm:w-72">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search registered models..."
              className="h-10 pl-9 text-sm bg-background/50 border-border/60"
            />
          </div>
        </div>

        {/* Model rows */}
        {filteredModels.length === 0 ? (
          <div className="rounded-xl border border-dashed border-border p-12 text-center bg-card">
            <Database className="mx-auto h-10 w-10 text-muted-foreground opacity-50" />
            <h3 className="mt-4 text-base font-semibold text-foreground">
              No models found
            </h3>
            <p className="mt-1 text-xs text-muted-foreground max-w-sm mx-auto">
              No models matched "{searchQuery}". Try clearing the search query.
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-4 text-xs"
              onClick={() => setSearchQuery("")}
            >
              Clear filter
            </Button>
          </div>
        ) : (
          <Card className="overflow-hidden">
            <ul className="divide-y divide-border/60">
              {filteredModels.map((model) => {
                const caps = [
                  model.permissions.view && "Read",
                  model.permissions.add && "Create",
                  model.permissions.change && "Update",
                  model.permissions.delete && "Delete",
                ].filter(Boolean);
                return (
                  <li
                    key={model.name}
                    className="group flex items-center gap-3 px-4 py-3 hover:bg-muted/40 transition-colors"
                  >
                    <span
                      aria-hidden
                      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-border/60 bg-muted/50 text-muted-foreground group-hover:text-primary group-hover:border-primary/30 group-hover:bg-primary/[0.07] transition-colors"
                    >
                      <ModelIcon name={model.icon} className="h-4 w-4" />
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-baseline gap-2">
                        <span className="font-semibold text-ui truncate">
                          {model.verbose_name_plural}
                        </span>
                        <span className="shrink-0 text-micro tabular-nums text-muted-foreground">
                          {model.count} {model.count === 1 ? "record" : "records"}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground truncate">
                        <span className="font-mono">{model.name}</span>
                        {caps.length > 0 && <span> · {caps.join(" · ")}</span>}
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
                            params: { model: model.name },
                          })
                        }
                      >
                        Browse
                        <ArrowRight className="h-3.5 w-3.5 ml-1" />
                      </Button>
                      {model.permissions.add && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 text-xs gap-1 px-2.5"
                          onClick={() =>
                            navigate({
                              to: "/$model/create",
                              params: { model: model.name },
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
    </AdminLayout>
  );
}
