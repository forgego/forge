import { useParams, useNavigate } from "@tanstack/react-router";
import { useQueryClient, useMutation } from "@tanstack/react-query";
import {
  useModelMetadata,
  useModelList,
  useDeleteObject,
  useBulkDelete,
  useDeleteSavedView,
  adminKeys,
  useSavedViews,
  useSaveSavedView,
} from "../api/hooks/adminHooks";
import { adminAPI } from "../api/client";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../components/ui/table";
import { Button } from "../components/ui/button";
import { Checkbox } from "../components/ui/checkbox";
import { Card, CardContent, CardHeader } from "../components/ui/card";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "../components/ui/dropdown-menu";
import {
  Loader2,
  Plus,
  Search,
  Download,
  FileSpreadsheet,
  FileCode,
  ChevronDown,
} from "lucide-react";
import { useState, useMemo } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import { useUIComponent } from "../hooks/useUIComponent";
import { useDebouncedValue } from "../hooks/useDebouncedValue";
import { cn } from "../lib/utils";
import { useToast } from "../hooks/use-toast";
import { ConfirmationDialog } from "../components/ui/confirmation-dialog";
import { PageHeader } from "../components/ui/page-header";
import { ListFilterPanel } from "../components/list/ListFilterPanel";
import { ListBulkToolbar } from "../components/list/ListBulkToolbar";
import { ListCell } from "../components/list/ListCell";
import { ListToolbar } from "../components/list/ListToolbar";
import { ListPagination } from "../components/list/ListPagination";
import { ListTableHeader } from "../components/list/ListTableHeader";
import { ListRowActions } from "../components/list/ListRowActions";
import { SaveViewDialog } from "../components/list/SaveViewDialog";

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

export default function ModelListPage() {
  const params = useParams({ strict: false }) as any;
  const { model } = params;
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const [searchInput, setSearchInput] = useState("");
  const debouncedSearch = useDebouncedValue(searchInput, 350);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState<number | null>(null);
  const [selectedIds, setSelectedIds] = useState<(string | number)[]>([]);
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const [sortField, setSortField] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<"asc" | "desc" | null>(null);
  const [selectedSavedView, setSelectedSavedView] = useState<string>("");
  const [isFilterOpen, setIsFilterOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | number | null>(null);
  const [bulkDeleteConfirmOpen, setBulkDeleteConfirmOpen] = useState(false);
  const [bulkAction, setBulkAction] = useState<any | null>(null);
  const [saveViewOpen, setSaveViewOpen] = useState(false);
  const [saveViewName, setSaveViewName] = useState("");
  const [exporting, setExporting] = useState(false);

  const modelName = model as string;

  const { data: metadata, isLoading: metaLoading } =
    useModelMetadata(modelName);

  const effectivePageSize =
    pageSize ?? metadata?.pagination?.page_size ?? 20;
  const maxPageSize = metadata?.pagination?.max_page_size ?? 100;
  // Always include the server default so the select never shows a
  // mismatched value.
  const pageSizeOptions = Array.from(
    new Set([effectivePageSize, ...PAGE_SIZE_OPTIONS])
  )
    .filter((s) => s <= maxPageSize)
    .sort((a, b) => a - b);

  const ordering =
    sortField && sortDir
      ? sortDir === "desc"
        ? `-${sortField}`
        : sortField
      : undefined;

  const {
    data: listData,
    isLoading: listLoading,
    error,
    refetch,
  } = useModelList(modelName, {
    page,
    page_size: effectivePageSize,
    search: debouncedSearch || undefined,
    ordering,
    ...activeFilters,
  });

  const deleteMutation = useDeleteObject(modelName);
  const bulkDeleteMutation = useBulkDelete(modelName);
  const deleteSavedViewMutation = useDeleteSavedView(modelName);
  const { data: savedViewsData } = useSavedViews(modelName);

  // UI override slots. Called unconditionally (before any early return)
  // so hook order stays stable; they render only after metadata loads.
  const ListTitle = useUIComponent(
    metadata?.ui_overrides?.["list.title"] || "",
    "div"
  );
  const ActionHeader = useUIComponent(
    metadata?.ui_overrides?.["list.actions"] || "",
    "div"
  );
  const ListItem = useUIComponent(
    metadata?.ui_overrides?.["list.row"] || "",
    "tr"
  );
  const { mutateAsync: saveView, isPending: saveViewPending } =
    useSaveSavedView(modelName, {
      onSuccess: (view) => {
        setSelectedSavedView(view.id);
        setSaveViewOpen(false);
        setSaveViewName("");
        toast({
          title: "View saved",
          description: `"${view.name}" is now available in Saved views.`,
        });
      },
      onError: (err: any) => {
        toast({
          title: "Could not save view",
          description: err.message || "Failed to save view",
          variant: "destructive",
        });
      },
    });

  // Bulk action mutation
  const { mutateAsync: runAction, isPending: actionLoading } = useMutation({
    mutationFn: ({ action, ids }: { action: string; ids: (string | number)[] }) =>
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

  const resetPage = () => setPage(1);

  const relationByField = useMemo(() => {
    const map = new Map<string, any>();
    (metadata?.relations ?? []).forEach((rel: any) => {
      if (!rel?.name) return;
      map.set(rel.name, rel);
      // A column is often the raw fk column, e.g. relation "customer" -> column "customer_id"
      map.set(`${rel.name}_id`, rel);
    });
    return map;
  }, [metadata]);

  const handleSearchChange = (value: string) => {
    setSearchInput(value);
    resetPage();
  };

  const clearSearchAndFilters = () => {
    setSearchInput("");
    setActiveFilters({});
    setSortField(null);
    setSortDir(null);
    setSelectedSavedView("");
    resetPage();
  };

  const handleToggleSort = (fieldName: string) => {
    if (sortField !== fieldName) {
      setSortField(fieldName);
      setSortDir("asc");
    } else if (sortDir === "asc") {
      setSortDir("desc");
    } else if (sortDir === "desc") {
      setSortField(null);
      setSortDir(null);
    }
    resetPage();
  };

  const handleDelete = async () => {
    if (deleteId === null) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      setSelectedIds((prev) =>
        prev.filter((selectedId) => selectedId !== deleteId)
      );
      toast({
        title: "Delete Successful",
        description: "The item has been removed.",
      });
    } catch (err: any) {
      toast({
        title: "Delete failed",
        description: err.message || "The item could not be removed.",
        variant: "destructive",
      });
    } finally {
      setDeleteId(null);
    }
  };

  const handleBulkDeleteConfirm = async () => {
    try {
      const result = await bulkDeleteMutation.mutateAsync(selectedIds);
      const total = selectedIds.length;
      const deleted = result?.deleted ?? total;
      const failed = result?.errors?.length ?? 0;
      setSelectedIds([]);
      if (failed > 0) {
        toast({
          title: "Partial delete",
          description: `Deleted ${deleted} of ${total} records. ${failed} failed.`,
          variant: "destructive",
        });
      } else {
        toast({
          title: "Bulk Delete Successful",
          description: `Successfully removed ${deleted} records.`,
        });
      }
    } catch (err: any) {
      toast({
        title: "Bulk Delete Failed",
        description: err.message || "Failed to delete records.",
        variant: "destructive",
      });
    } finally {
      setBulkDeleteConfirmOpen(false);
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(listData?.results?.map((obj: any) => obj.id) || []);
    } else {
      setSelectedIds([]);
    }
  };

  const toggleSelect = (id: string | number) => {
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

  const exportParams = () => {
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
    return {
      search: debouncedSearch || undefined,
      ordering,
      ...filteredParams,
    };
  };

  const handleExportChange = async (format: "csv" | "json") => {
    setExporting(true);
    try {
      const { blob, filename } = await adminAPI.downloadExport(
        modelName,
        format,
        exportParams()
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      toast({
        title: "Export ready",
        description: `${filename} has been downloaded.`,
      });
    } catch (err: any) {
      toast({
        title: "Export failed",
        description: err.message || "Could not export records.",
        variant: "destructive",
      });
    } finally {
      setExporting(false);
    }
  };

  const handleCreateModel = () =>
    navigate({
      to: "/$model/create",
      params: { model: modelName },
    });

  const handleSaveViewConfirm = async () => {
    const name = saveViewName.trim();
    if (!name) return;
    await saveView({
      name,
      filters: activeFilters,
      ordering: ordering ? [ordering] : [],
      display: displayFields,
    });
  };

  const handleApplySavedView = (viewId: string) => {
    setSelectedSavedView(viewId);
    const view = savedViews.find((item) => item.id === viewId);
    if (!view) {
      setActiveFilters({});
      setSortField(null);
      setSortDir(null);
      resetPage();
      return;
    }
    setActiveFilters(view.filters || {});
    const firstOrdering = view.ordering?.[0];
    if (firstOrdering) {
      setSortDir(firstOrdering.startsWith("-") ? "desc" : "asc");
      setSortField(firstOrdering.replace(/^-/, ""));
    } else {
      setSortField(null);
      setSortDir(null);
    }
    resetPage();
  };

  const handleDeleteSavedView = async () => {
    if (!selectedSavedView) return;
    try {
      await deleteSavedViewMutation.mutateAsync(selectedSavedView);
      setSelectedSavedView("");
      toast({ title: "View deleted" });
    } catch (err: any) {
      toast({
        title: "Could not delete view",
        description: err.message || "Failed to delete view",
        variant: "destructive",
      });
    }
  };

  if (metaLoading || (listLoading && !listData)) {
    return (
      <AdminLayout>
        <div
          className="flex flex-col items-center justify-center h-96 gap-3"
          role="status"
          aria-label="Loading records"
        >
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-ui text-muted-foreground">Loading records…</p>
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-96">
          <div className="text-center space-y-3">
            <h2 className="text-title font-bold text-destructive">
              Could not load records
            </h2>
            <p className="text-muted-foreground mt-2 max-w-md">
              {(error as Error).message}
            </p>
            <Button
              data-testid="list-retry"
              variant="outline"
              onClick={() => refetch()}
            >
              Retry
            </Button>
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
  const totalCount = listData.count ?? 0;
  const totalPages = listData.total_pages || 1;
  const from = totalCount === 0 ? 0 : (page - 1) * effectivePageSize + 1;
  const to = Math.min(page * effectivePageSize, totalCount);
  const activeFilterCount = Object.values(activeFilters).filter(
    (v) => v !== "" && v !== null && v !== undefined
  ).length;
  const allSelected =
    objects.length > 0 && selectedIds.length === objects.length;
  const someSelected =
    selectedIds.length > 0 && selectedIds.length < objects.length;
  const hasActiveSearchOrFilters =
    debouncedSearch !== "" || activeFilterCount > 0;

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* eslint-disable-next-line react-hooks/static-components -- resolved via useUIComponent registry (stable ref) */}
        <ListTitle>
          <PageHeader
            eyebrow="Records"
            title={metadata.verbose_name_plural}
            description={metadata.description || undefined}
            actions={
              /* eslint-disable-next-line react-hooks/static-components -- resolved via useUIComponent registry (stable ref) */
              <ActionHeader className="flex items-center gap-2">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      data-testid="export-select"
                      variant="outline"
                      className="h-10 px-4 text-body font-medium gap-2"
                      disabled={exporting}
                    >
                      {exporting ? (
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                      ) : (
                        <Download className="h-4 w-4 text-muted-foreground" />
                      )}
                      <span>Export</span>
                      <ChevronDown className="h-3 w-3 text-muted-foreground" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-44">
                    <DropdownMenuItem
                      data-testid="export-csv"
                      className="gap-2 cursor-pointer text-meta"
                      onClick={() => handleExportChange("csv")}
                    >
                      <FileSpreadsheet className="h-4 w-4 text-success" />
                      <span>Export CSV</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      data-testid="export-json"
                      className="gap-2 cursor-pointer text-meta"
                      onClick={() => handleExportChange("json")}
                    >
                      <FileCode className="h-4 w-4 text-info" />
                      <span>Export JSON</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                {metadata.permissions.add && (
                  <Button
                    data-testid="create-button"
                    onClick={handleCreateModel}
                    className="bg-primary hover:bg-primary/90 h-10"
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    Add {metadata.verbose_name}
                  </Button>
                )}
              </ActionHeader>
            }
            meta={
              <span>
                <span className="font-mono tabular-nums font-semibold">{totalCount.toLocaleString()}</span>{" "}
                {totalCount === 1 ? "item" : "items"}
              </span>
            }
          />
        </ListTitle>

        <Card className="border-border-subtle overflow-hidden">
          <CardHeader className="bg-muted/40 border-b border-border-subtle py-4 space-y-4">
            <ListToolbar
              searchLabel={metadata.verbose_name_plural}
              searchInput={searchInput}
              onSearchChange={handleSearchChange}
              savedViews={savedViews}
              selectedSavedView={selectedSavedView}
              onApplySavedView={handleApplySavedView}
              onOpenSaveView={() => {
                setSaveViewName("");
                setSaveViewOpen(true);
              }}
              onDeleteSavedView={handleDeleteSavedView}
              deleteViewPending={deleteSavedViewMutation.isPending}
              isFilterOpen={isFilterOpen}
              onToggleFilters={() => setIsFilterOpen(!isFilterOpen)}
              activeFilterCount={activeFilterCount}
            />

            {/* Expansible Filter Panel */}
            {isFilterOpen && metadata.filters.length > 0 && (
              <ListFilterPanel
                filters={metadata.filters}
                activeFilters={activeFilters}
                setActiveFilters={setActiveFilters}
                resetPage={resetPage}
              />
            )}

            <ListBulkToolbar
              hasSelection={hasSelection}
              selectedIds={selectedIds}
              setSelectedIds={setSelectedIds}
              actions={metadata.actions}
              permissions={metadata.permissions}
              handleActionClick={handleActionClick}
              actionLoading={actionLoading}
              setBulkDeleteConfirmOpen={setBulkDeleteConfirmOpen}
              bulkDeletePending={bulkDeleteMutation.isPending}
            />
          </CardHeader>
          <CardContent
            className="p-0 overflow-x-auto [background:linear-gradient(to_right,hsl(var(--surface-2))_30%,transparent),linear-gradient(to_right,transparent,hsl(var(--surface-2))_70%)_right,radial-gradient(farthest-side_at_0_50%,rgba(0,0,0,0.12),transparent),radial-gradient(farthest-side_at_100%_50%,rgba(0,0,0,0.12),transparent)_right] [background-repeat:no-repeat] [background-size:40px_100%,40px_100%,14px_100%,14px_100%] [background-attachment:local,local,scroll,scroll]"
            style={{ ["--sticky-id-offset" as any]: "3rem" }}
          >
            <Table className="tabular-nums">
              <ListTableHeader
                allSelected={allSelected}
                someSelected={someSelected}
                onSelectAll={handleSelectAll}
                displayFields={displayFields}
                fieldsByName={fieldsByName}
                sortField={sortField}
                sortDir={sortDir}
                onToggleSort={handleToggleSort}
              />
              <TableBody>
                {objects.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={displayFields.length + 2}
                      className="text-center py-16 text-muted-foreground"
                    >
                      <div className="flex flex-col items-center gap-2 max-w-sm mx-auto">
                        <Search className="h-10 w-10 opacity-20" />
                        {hasActiveSearchOrFilters ? (
                          <>
                            <p className="text-ui font-medium text-foreground">
                              No results found
                            </p>
                            <p className="text-meta">
                              Try adjusting your search or filters.
                            </p>
                            <Button
                              variant="outline"
                              size="sm"
                              data-testid="clear-filters"
                              className="mt-2"
                              onClick={clearSearchAndFilters}
                            >
                              Clear search & filters
                            </Button>
                          </>
                        ) : (
                          <>
                            <p className="text-ui font-medium text-foreground">
                              No {metadata.verbose_name_plural.toLowerCase()} yet
                            </p>
                            {metadata.permissions.add && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="mt-2"
                                onClick={handleCreateModel}
                              >
                                <Plus className="h-3.5 w-3.5 mr-1.5" />
                                Add {metadata.verbose_name}
                              </Button>
                            )}
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  objects.map((obj: any) => {
                    const isSelected = selectedIds.includes(obj.id);
                    return (
                    <ListItem
                      key={obj.id}
                      className={cn(
                        "group hover:bg-surface-sunken/60 border-border-subtle transition-colors",
                        isSelected &&
                          "bg-primary/5 hover:bg-primary/10"
                      )}
                    >
                      <TableCell
                        className={cn(
                          "pl-6 py-3 sticky left-0 z-20",
                          isSelected ? "bg-inherit" : "bg-surface-2"
                        )}
                      >
                        <Checkbox
                          data-testid={`select-${obj.id}`}
                          aria-label={`Select row ${obj.id}`}
                          checked={selectedIds.includes(obj.id)}
                          onChange={() => toggleSelect(obj.id)}
                        />
                      </TableCell>
                      {displayFields.map((fieldName, colIdx) => (
                        <ListCell
                          key={fieldName}
                          fieldName={fieldName}
                          colIdx={colIdx}
                          obj={obj}
                          field={fieldsByName.get(fieldName)}
                          relation={relationByField.get(fieldName)}
                          navigate={navigate}
                          isSelected={isSelected}
                          metadata={metadata}
                          modelName={modelName}
                          display={listData?.display}
                        />
                      ))}
                      <ListRowActions
                        id={obj.id}
                        modelName={modelName}
                        permissions={metadata.permissions}
                        navigate={navigate}
                        onDelete={() => setDeleteId(obj.id)}
                        deletePending={deleteMutation.isPending}
                      />
                    </ListItem>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </CardContent>
          <ListPagination
            from={from}
            to={to}
            totalCount={totalCount}
            page={page}
            totalPages={totalPages}
            pageSize={effectivePageSize}
            pageSizeOptions={pageSizeOptions}
            onPageSizeChange={(v) => {
              setPageSize(Number(v));
              resetPage();
            }}
            onPrev={() => setPage((p) => Math.max(1, p - 1))}
            onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
          />
        </Card>

        {/* Save view dialog */}
        <SaveViewDialog
          open={saveViewOpen}
          onOpenChange={setSaveViewOpen}
          name={saveViewName}
          onNameChange={setSaveViewName}
          onConfirm={handleSaveViewConfirm}
          pending={saveViewPending}
        />

        {/* Dialogs */}
        <ConfirmationDialog
          open={deleteId !== null}
          onOpenChange={(open) => !open && setDeleteId(null)}
          testId="delete-dialog"
          title="Delete Item"
          description="Are you sure you want to delete this item? This action cannot be undone."
          confirmLabel="Delete"
          variant="destructive"
          onConfirm={handleDelete}
        />

        <ConfirmationDialog
          open={!!bulkAction}
          onOpenChange={(open) => !open && setBulkAction(null)}
          testId="bulk-action-dialog"
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

        <ConfirmationDialog
          open={bulkDeleteConfirmOpen}
          onOpenChange={setBulkDeleteConfirmOpen}
          testId="bulk-delete-dialog"
          title={`Delete ${selectedIds.length} ${selectedIds.length === 1 ? metadata.verbose_name : metadata.verbose_name_plural}`}
          description={`Are you sure you want to delete ${selectedIds.length} selected records? This action cannot be undone.`}
          confirmLabel="Delete Selected"
          variant="destructive"
          onConfirm={handleBulkDeleteConfirm}
        />
      </div>
    </AdminLayout>
  );
}
