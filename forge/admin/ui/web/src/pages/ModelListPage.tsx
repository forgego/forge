import { useParams, useNavigate } from "@tanstack/react-router";
import { useQueryClient, useMutation } from "@tanstack/react-query";
import {
  useModelMetadata,
  useModelList,
  useDeleteObject,
  adminKeys,
  useSavedViews,
  useSaveSavedView,
} from "../api/hooks/adminHooks";
import { adminAPI } from "../api/client";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card, CardContent, CardHeader } from "../components/ui/card";
import {
  Loader2,
  Plus,
  Search,
  Edit,
  Trash2,
  Filter,
  CheckCircle,
  XCircle,
  Download,
} from "lucide-react";
import { useState } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import { useUIComponent } from "../hooks/useUIComponent";
import { cn } from "../lib/utils";
import { useToast } from "../hooks/use-toast";
import { ConfirmationDialog } from "../components/ui/confirmation-dialog";

export default function ModelListPage() {
  const params = useParams({ strict: false }) as any;
  const { model } = params;
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const [ordering, setOrdering] = useState<string[]>([]);
  const [selectedSavedView, setSelectedSavedView] = useState<string>("");
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [bulkAction, setBulkAction] = useState<any | null>(null);
  const [exportFormat, setExportFormat] = useState("");

  const modelName = model as string;

  const { data: metadata, isLoading: metaLoading } =
    useModelMetadata(modelName);
  const {
    data: listData,
    isLoading: listLoading,
    error,
  } = useModelList(modelName, {
    page,
    search: searchQuery || undefined,
    ordering: ordering.length ? ordering.join(",") : undefined,
    ...activeFilters,
  });

  const deleteMutation = useDeleteObject(modelName);
  const { data: savedViewsData } = useSavedViews(modelName);
  const { mutateAsync: saveView, isPending: saveViewPending } =
    useSaveSavedView(modelName, {
      onSuccess: (view) => {
        setSelectedSavedView(view.id);
        toast({
          title: "Saved view updated",
          description: `Saved view \"${view.name}\" is ready to use.`,
        });
      },
      onError: (err: any) => {
        toast({
          title: "Error",
          description: err.message || "Failed to save view",
          variant: "destructive",
        });
      },
    });

  // Bulk action mutation
  const { mutateAsync: runAction, isPending: actionLoading } = useMutation({
    mutationFn: ({ action, ids }: { action: string; ids: number[] }) =>
      adminAPI.executeAction(modelName, action, { ids }),
    onSuccess: (data) => {
      toast({
        title: "Success",
        description: data.message || "Action executed successfully",
      });
      setSelectedIds([]);
      setBulkAction(null);
      queryClient.invalidateQueries({ queryKey: adminKeys.model(modelName) });
    },
    onError: (err: any) => {
      toast({
        title: "Error",
        description: err.message || "Failed to execute action",
        variant: "destructive",
      });
      setBulkAction(null);
    },
  });

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      setSelectedIds((prev) =>
        prev.filter((selectedId) => selectedId !== deleteId)
      );
      toast({
        title: "Delete Successful",
        description: "The item has been removed.",
      });
    } catch (error) {
      console.error("Failed to delete:", error);
    } finally {
      setDeleteId(null);
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(listData?.results?.map((obj: any) => obj.id) || []);
    } else {
      setSelectedIds([]);
    }
  };

  const toggleSelect = (id: number) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]
    );
  };

  const handleActionClick = (action: any) => {
    if (action.confirmation) {
      setBulkAction(action);
    } else {
      runAction({ action: action.name, ids: selectedIds });
    }
  };
  const hasSelection = selectedIds.length > 0;

  const handleExportChange = (format: string) => {
    if (!format) return;
    const filteredParams = Object.entries(activeFilters).reduce(
      (acc, [key, value]) => {
        if (value === "" || value === null || value === undefined) {
          return acc;
        }
        acc[key] = value;
        return acc;
      },
      {} as Record<string, any>
    );

    const exportUrl = adminAPI.getExportURL(
      modelName,
      format as "csv" | "json",
      {
        search: searchQuery || undefined,
        ...filteredParams,
      }
    );
    window.open(exportUrl, "_blank", "noopener,noreferrer");
    setExportFormat("");
  };

  if (metaLoading || (listLoading && !listData)) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-destructive">Error</h2>
            <p className="text-muted-foreground mt-2">{error.message}</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  if (!metadata || !listData) return null;

  const displayFields = metadata.list_display || [];
  const fieldsByName = new Map(
    metadata.fields.map((field) => [field.name, field])
  );
  const objects = listData.results || [];
  const savedViews = savedViewsData?.views || [];

  // Resolve Overrides
  const ListTitle = useUIComponent(
    metadata.ui_overrides?.["list.title"] || "",
    "div"
  );
  const ActionHeader = useUIComponent(
    metadata.ui_overrides?.["list.actions"] || "",
    "div"
  );
  const ListItem = useUIComponent(
    metadata.ui_overrides?.["list.row"] || "",
    "tr"
  );

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <ListTitle className="space-y-1.5">
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">
              {metadata.verbose_name_plural}
            </h1>
            <p className="text-sm text-muted-foreground max-w-2xl">
              {metadata.description}
            </p>
          </ListTitle>
          <ActionHeader className="flex items-center gap-2">
            <div className="relative">
              <select
                data-testid="export-select"
                className="h-11 pl-3 pr-8 text-xs font-semibold uppercase tracking-wider bg-background/50 border border-border/50 rounded-md text-muted-foreground hover:text-foreground focus:outline-none focus:ring-2 focus:ring-primary/30"
                value={exportFormat}
                onChange={(e) => {
                  const value = e.target.value;
                  setExportFormat(value);
                  handleExportChange(value);
                }}
              >
                <option value="">Export</option>
                <option value="csv">Export CSV</option>
                <option value="json">Export JSON</option>
              </select>
              <Download className="pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            </div>
            {metadata.permissions.add && (
              <Button
                data-testid="create-button"
                onClick={() =>
                  navigate({
                    to: "/$model/create",
                    params: { model: modelName },
                  })
                }
                className="bg-primary hover:bg-primary/90"
              >
                <Plus className="h-4 w-4 mr-2" />
                Add {metadata.verbose_name}
              </Button>
            )}
          </ActionHeader>
        </div>

        <Card className="glass-lite border-border/50 shadow-xl shadow-black/5 overflow-hidden">
          <CardHeader className="bg-muted/50 border-b border-border/50 py-4">
            <div className="flex flex-col sm:flex-row items-center gap-4">
              <div className="relative flex-1 w-full">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  data-testid="search-input"
                  placeholder={`Search ${metadata.verbose_name_plural.toLowerCase()}...`}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-background/50 border-border/50 focus:ring-primary/20 h-11"
                />
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <div className="flex items-center gap-2">
                  <div className="h-11 px-3 flex items-center bg-background/50 border border-border/50 rounded-md">
                    <select
                      className="bg-transparent text-xs font-medium uppercase tracking-wide text-muted-foreground focus:outline-none"
                      value={selectedSavedView}
                      onChange={(e) => {
                        const value = e.target.value;
                        setSelectedSavedView(value);
                        const view = savedViews.find((item) => item.id === value);
                        if (!view) {
                          setActiveFilters({});
                          setOrdering([]);
                          setPage(1);
                          return;
                        }
                        setActiveFilters(view.filters || {});
                        setOrdering(view.ordering || []);
                        setPage(1);
                      }}
                    >
                      <option value="">Saved Views</option>
                      {savedViews.map((view) => (
                        <option key={view.id} value={view.id}>
                          {view.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-11 px-4 text-[10px] font-bold uppercase tracking-widest"
                    onClick={async () => {
                      const name = window.prompt("Name this view");
                      if (!name) {
                        return;
                      }
                      await saveView({
                        name,
                        filters: activeFilters,
                        ordering,
                        display: displayFields,
                      });
                    }}
                    disabled={saveViewPending}
                  >
                    {saveViewPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      "Save current view"
                    )}
                  </Button>
                </div>
                <Button
                  data-testid="filter-button"
                  variant="ghost"
                  size="sm"
                  className={cn(
                    "h-11 px-4 text-xs font-semibold uppercase tracking-widest bg-background/40 text-muted-foreground hover:text-foreground",
                    isFilterOpen &&
                      "bg-primary/10 text-primary"
                  )}
                  onClick={() => setIsFilterOpen(!isFilterOpen)}
                >
                  <Filter className="h-4 w-4 mr-2" />
                  Filters
                </Button>
                <div className="h-11 px-4 flex items-center bg-background/50 border border-border/50 rounded-md">
                  <span className="text-xs font-bold uppercase tracking-widest text-muted-foreground">
                    {listData.count}{" "}
                    <span className="font-normal opacity-70 italic lowercase">
                      items
                    </span>
                  </span>
                </div>
              </div>
            </div>

            {/* Expansible Filter Panel */}
            {isFilterOpen && metadata.filters.length > 0 && (
              <div className="mt-4 p-4 bg-background/60 rounded-lg border border-border/50 animate-in slide-in-from-top-2 duration-300">
                <div className="flex items-center justify-between gap-2 mb-4">
                  <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                    Filters
                  </p>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground hover:text-foreground"
                    onClick={() => setActiveFilters({})}
                  >
                    Reset filters
                  </Button>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
                  {metadata.filters.map((filter) => (
                    <div key={filter.name} className="space-y-1.5">
                      <label className="text-[10px] font-bold uppercase tracking-tighter text-muted-foreground/80">
                        {filter.label}
                      </label>
                      {filter.type === "choice" ? (
                        <select
                          data-testid={`filter-${filter.name}`}
                          className="w-full bg-background/80 border border-border/50 rounded-md px-3 py-1.5 text-xs focus:ring-1 focus:ring-primary/30 outline-none"
                          value={activeFilters[filter.name] || ""}
                          onChange={(e) =>
                            setActiveFilters((prev) => ({
                              ...prev,
                              [filter.name]: e.target.value,
                            }))
                          }
                        >
                          <option value="">All</option>
                          {filter.choices?.map((c) => (
                            <option
                              key={String(c.value)}
                              value={String(c.value)}
                            >
                              {c.label}
                            </option>
                          ))}
                        </select>
                      ) : filter.type === "date" ||
                        filter.type === "datetime" ? (
                        <div className="flex gap-2">
                          <Input
                            type="date"
                            placeholder="Start"
                            className="h-8 text-xs bg-background/80"
                            value={activeFilters[`${filter.name}__gte`] || ""}
                            onChange={(e) =>
                              setActiveFilters((prev) => ({
                                ...prev,
                                [`${filter.name}__gte`]: e.target.value,
                              }))
                            }
                          />
                          <Input
                            type="date"
                            placeholder="End"
                            className="h-8 text-xs bg-background/80"
                            value={activeFilters[`${filter.name}__lte`] || ""}
                            onChange={(e) =>
                              setActiveFilters((prev) => ({
                                ...prev,
                                [`${filter.name}__lte`]: e.target.value,
                              }))
                            }
                          />
                        </div>
                      ) : filter.type === "boolean" ? (
                        <select
                          data-testid={`filter-${filter.name}`}
                          className="w-full bg-background/80 border border-border/50 rounded-md px-3 py-1.5 text-xs focus:ring-1 focus:ring-primary/30 outline-none"
                          value={activeFilters[filter.name] || ""}
                          onChange={(e) =>
                            setActiveFilters((prev) => ({
                              ...prev,
                              [filter.name]: e.target.value,
                            }))
                          }
                        >
                          <option value="">All</option>
                          <option value="true">Yes</option>
                          <option value="false">No</option>
                        </select>
                      ) : (
                        <Input
                          data-testid={`filter-${filter.name}`}
                          placeholder={filter.label}
                          className="h-8 text-xs bg-background/80"
                          value={activeFilters[filter.name] || ""}
                          onChange={(e) =>
                            setActiveFilters((prev) => ({
                              ...prev,
                              [filter.name]: e.target.value,
                            }))
                          }
                        />
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div
              data-testid="bulk-toolbar"
              className={cn(
                "mt-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 rounded-lg border border-border/40 px-4 py-3 text-xs",
                hasSelection
                  ? "bg-primary/5 text-foreground"
                  : "bg-background/40 text-muted-foreground"
              )}
            >
              <div className="flex flex-wrap items-center gap-3">
                <span className="font-semibold uppercase tracking-widest">
                  {hasSelection
                    ? `${selectedIds.length} selected`
                    : "Select rows to enable bulk actions"}
                </span>
                <div className="h-4 w-px bg-border/60 hidden sm:block" />
                <div className="flex flex-wrap items-center gap-2">
                  {metadata.actions.map((action) => (
                    <Button
                      key={action.name}
                      data-testid={`bulk-action-${action.name}`}
                      variant="ghost"
                      size="sm"
                      className={cn(
                        "h-8 px-3 text-[10px] font-semibold uppercase tracking-widest",
                        !hasSelection &&
                          "text-muted-foreground/60 hover:text-muted-foreground"
                      )}
                      onClick={() => handleActionClick(action)}
                      disabled={!hasSelection || actionLoading}
                    >
                      {actionLoading ? (
                        <Loader2 className="h-3 w-3 animate-spin mr-2" />
                      ) : (
                        action.icon && <span className="mr-2">⚡</span>
                      )}
                      {action.label}
                    </Button>
                  ))}
                </div>
              </div>
              <Button
                data-testid="bulk-cancel"
                variant="ghost"
                size="sm"
                className="h-8 px-3 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground hover:text-foreground"
                onClick={() => setSelectedIds([])}
                disabled={!hasSelection}
              >
                Clear selection
              </Button>
            </div>
          </CardHeader>
          <CardContent className="p-0 overflow-x-auto">
            <Table>
              <TableHeader className="bg-muted/30">
                <TableRow className="hover:bg-transparent border-border/50">
                  <TableHead className="w-[50px] pl-6 py-4">
                    <input
                      type="checkbox"
                      data-testid="select-all"
                      className="h-4 w-4 rounded border-border bg-background text-primary focus:ring-primary/20"
                      checked={
                        selectedIds.length === objects.length &&
                        objects.length > 0
                      }
                      onChange={(e) => handleSelectAll(e.target.checked)}
                    />
                  </TableHead>
                  {displayFields.map((fieldName) => {
                    const field = fieldsByName.get(fieldName);
                    return (
                      <TableHead
                        key={fieldName}
                        className="font-bold text-[11px] uppercase tracking-wider text-muted-foreground py-4 min-w-[140px]"
                      >
                        {field?.label || fieldName}
                      </TableHead>
                    );
                  })}
                  <TableHead className="w-[80px] font-bold text-[11px] uppercase tracking-wider text-muted-foreground text-center">
                    Actions
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {objects.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={displayFields.length + 2}
                      className="text-center py-20 text-muted-foreground"
                    >
                      <div className="flex flex-col items-center gap-2">
                        <Search className="h-10 w-10 opacity-20" />
                        <p className="text-sm font-medium">
                          No results found matching your criteria
                        </p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  objects.map((obj: any) => (
                    <ListItem
                      key={obj.id}
                      className={cn(
                        "group hover:bg-muted/30 border-border/50 transition-colors",
                        selectedIds.includes(obj.id) &&
                          "bg-primary/5 hover:bg-primary/10"
                      )}
                    >
                      <TableCell className="pl-6 py-4">
                        <input
                          type="checkbox"
                          data-testid={`select-${obj.id}`}
                          className="h-4 w-4 rounded border-border bg-background text-primary focus:ring-primary/20"
                          checked={selectedIds.includes(obj.id)}
                          onChange={() => toggleSelect(obj.id)}
                        />
                      </TableCell>
                      {displayFields.map((fieldName) => (
                        <TableCell
                          key={fieldName}
                          className="py-4 text-sm font-medium text-foreground/80"
                        >
                          {fieldName === "active" ||
                          typeof obj[fieldName] === "boolean" ? (
                            obj[fieldName] ? (
                              <CheckCircle className="h-4 w-4 text-emerald-500" />
                            ) : (
                              <XCircle className="h-4 w-4 text-rose-500" />
                            )
                          ) : (
                            obj[fieldName]?.toString() || (
                              <span className="opacity-30">—</span>
                            )
                          )}
                        </TableCell>
                      ))}
                      <TableCell className="text-center">
                        <div className="flex justify-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          {metadata.permissions.change && (
                            <Button
                              variant="ghost"
                              size="icon"
                              data-testid={`edit-${obj.id}`}
                              className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-primary/10"
                              onClick={() =>
                                navigate({
                                  to: "/$model/$id",
                                  params: { model: modelName, id: obj.id },
                                })
                              }
                            >
                              <Edit className="h-3.5 w-3.5" />
                            </Button>
                          )}
                          {metadata.permissions.delete && (
                            <Button
                              variant="ghost"
                              size="icon"
                              data-testid={`delete-${obj.id}`}
                              className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                              onClick={() => setDeleteId(obj.id)}
                              disabled={deleteMutation.isPending}
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </ListItem>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
          <div className="px-6 py-4 bg-muted/20 border-t border-border/50 flex flex-col sm:flex-row items-center justify-between gap-4">
            <p className="text-xs text-muted-foreground font-medium">
              Page <span className="text-foreground">{page}</span> of{" "}
              <span className="text-foreground">
                {listData.total_pages || 1}
              </span>
            </p>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="h-9 px-4 text-xs font-bold uppercase tracking-tight bg-background/50 border-border/50 hover:bg-background"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="h-9 px-4 text-xs font-bold uppercase tracking-tight bg-background/50 border-border/50 hover:bg-background"
                onClick={() =>
                  setPage((p) => Math.min(listData.total_pages, p + 1))
                }
                disabled={
                  page === listData.total_pages || listData.total_pages === 0
                }
              >
                Next
              </Button>
            </div>
          </div>
        </Card>

        {/* Dialogs */}
        <ConfirmationDialog
          open={!!deleteId}
          onOpenChange={(open) => !open && setDeleteId(null)}
          title="Delete Item"
          description="Are you sure you want to delete this item? This action cannot be undone."
          confirmLabel="Delete"
          variant="destructive"
          onConfirm={handleDelete}
        />

        <ConfirmationDialog
          open={!!bulkAction}
          onOpenChange={(open) => !open && setBulkAction(null)}
          title={`Confirm ${bulkAction?.label}`}
          description={
            bulkAction?.confirmation || "Are you sure you want to proceed?"
          }
          confirmLabel="Confirm"
          onConfirm={() => {
            if (bulkAction) {
              runAction({ action: bulkAction.name, ids: selectedIds });
            }
          }}
        />
      </div>
    </AdminLayout>
  );
}
