import { useState, useMemo } from "react";
import { useNavigate, useParams, Link } from "@tanstack/react-router";
import {
  Loader2,
  ArrowLeft,
  Edit,
  Trash2,
  Copy,
  Check,
  ExternalLink,
  History,
  FileCode,
  Sliders,
  XCircle,
  Link2,
} from "lucide-react";
import AdminLayout from "../components/layout/AdminLayout";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Badge } from "../components/ui/badge";
import {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from "../components/ui/tabs";
import { ConfirmationDialog } from "../components/ui/confirmation-dialog";
import { AuditHistoryViewer } from "../components/history/AuditHistoryViewer";
import {
  useModelDetail,
  useModelMetadata,
  useModelHistory,
  useDeleteObject,
} from "../api/hooks/adminHooks";
import { useToast } from "../hooks/use-toast";
import { PageHeader } from "../components/ui/page-header";
import { EmptyValue } from "../components/ui/empty-state";
import { StatusBadge } from "../components/ui/status-badge";

function isDateField(name: string, type?: string): boolean {
  return (
    type === "date" ||
    type === "datetime" ||
    name.endsWith("_at") ||
    name.endsWith("_date") ||
    name === "created" ||
    name === "updated"
  );
}

function formatDisplayDate(val: unknown): string {
  try {
    const d = new Date(val as string | number);
    if (isNaN(d.getTime())) return String(val);
    return d.toLocaleString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return String(val);
  }
}

export default function ModelViewPage() {
  const params = useParams({ strict: false }) as Record<string, string>;
  const { model, id } = params;
  const navigate = useNavigate();
  const { toast } = useToast();

  const [activeTab, setActiveTab] = useState<string>("overview");
  const [copiedJSON, setCopiedJSON] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const modelName = model as string;
  const objectID = id as string;

  const { data: metadata, isLoading: metaLoading } = useModelMetadata(modelName);
  const {
    data: objectData,
    isLoading: objectLoading,
    error,
  } = useModelDetail(modelName, objectID, {
    enabled: Boolean(modelName && objectID),
  });

  const {
    data: historyData,
    isLoading: historyLoading,
  } = useModelHistory(modelName, objectID, {
    enabled: Boolean(modelName && objectID),
  });

  const deleteMutation = useDeleteObject(modelName);

  const relationByField = useMemo(() => {
    const map = new Map<string, any>();
    (metadata?.relations ?? []).forEach((rel: any) => {
      if (!rel?.name) return;
      map.set(rel.name, rel);
      map.set(`${rel.name}_id`, rel);
    });
    return map;
  }, [metadata]);

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(objectID);
      toast({
        title: "Record deleted",
        description: `${metadata?.verbose_name || "Record"} #${objectID} was successfully removed.`,
      });
      navigate({ to: "/$model", params: { model: modelName } });
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to delete record.";
      toast({
        title: "Deletion failed",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setDeleteDialogOpen(false);
    }
  };

  const copyJSON = async () => {
    if (!objectData) return;
    const payload = JSON.stringify(objectData, null, 2);
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(payload);
      } else {
        // Fallback for insecure contexts / older browsers.
        const area = document.createElement("textarea");
        area.value = payload;
        area.style.position = "fixed";
        area.style.opacity = "0";
        document.body.appendChild(area);
        area.select();
        document.execCommand("copy");
        area.remove();
      }
      setCopiedJSON(true);
      toast({
        title: "Copied to clipboard",
        description: "Full record JSON copied.",
      });
      setTimeout(() => setCopiedJSON(false), 2000);
    } catch {
      toast({
        title: "Copy failed",
        description: "Could not access the clipboard in this browser.",
        variant: "destructive",
      });
    }
  };

  if (metaLoading || objectLoading) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-ui text-muted-foreground animate-pulse">
            Loading record details...
          </p>
        </div>
      </AdminLayout>
    );
  }

  if (!metadata || error || !objectData) {
    return (
      <AdminLayout>
        <div className="flex flex-col items-center justify-center h-96 gap-4 text-center">
          <div className="rounded-full bg-destructive/10 p-4 text-destructive">
            <XCircle className="h-8 w-8" />
          </div>
          <div>
            <h2 className="text-title font-bold tracking-tight text-foreground">
              Unable to load record
            </h2>
            <p className="text-ui text-muted-foreground mt-1 max-w-md">
              Could not retrieve {metadata?.verbose_name || modelName} #{objectID}. The item may have been deleted or moved.
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => navigate({ to: "/$model", params: { model: modelName } })}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to {metadata?.verbose_name_plural || "records"}
          </Button>
        </div>
      </AdminLayout>
    );
  }

  const primaryLabel =
    objectData.name ||
    objectData.title ||
    objectData.label ||
    objectData.email ||
    objectData.username ||
    `#${objectID}`;

  // Check for status-like fields
  const statusField = metadata.fields.find(
    (f) =>
      f.name === "status" ||
      f.name === "state" ||
      f.name === "is_active" ||
      f.name === "active"
  );
  const statusValue = statusField ? objectData[statusField.name] : undefined;

  return (
    <AdminLayout>
      <div className="space-y-6 pb-12">
        <PageHeader
          eyebrow={metadata.verbose_name}
          title={primaryLabel}
          meta={
            <>
              <span className="font-mono tabular-nums">
                {metadata.name} #{objectID}
              </span>
              {statusValue !== undefined && (
                <StatusBadge
                  tone={
                    statusValue === true ||
                    statusValue === "active" ||
                    statusValue === "published" ||
                    statusValue === "completed" ||
                    statusValue === "enabled" ||
                    statusValue === "open" ||
                    statusValue === "paid" ||
                    statusValue === "approved"
                      ? "success"
                      : "neutral"
                  }
                >
                  {String(statusValue)}
                </StatusBadge>
              )}
            </>
          }
          actions={
            <>
              {metadata.permissions.change && (
                <Button
                  data-testid="edit-record"
                  onClick={() =>
                    navigate({
                      to: "/$model/$id",
                      params: { model: modelName, id: objectID },
                    })
                  }
                >
                  <Edit className="h-4 w-4 mr-2" />
                  Edit Record
                </Button>
              )}
              {metadata.permissions.delete && (
                <Button
                  data-testid="delete-record"
                  variant="outline"
                  className="border-destructive/30 text-destructive hover:bg-destructive/10"
                  onClick={() => setDeleteDialogOpen(true)}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete
                </Button>
              )}
            </>
          }
        />

        {/* Tab Navigation */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-3 sm:max-w-md">
            <TabsTrigger value="overview" data-testid="tab-overview" className="flex items-center gap-2 text-xs sm:text-sm">
              <Sliders className="h-3.5 w-3.5" />
              Overview
            </TabsTrigger>
            <TabsTrigger value="history" data-testid="tab-history" className="flex items-center gap-2 text-xs sm:text-sm">
              <History className="h-3.5 w-3.5" />
              Audit Log
              {historyData?.entries?.length ? (
                <span className="ml-1 rounded-full bg-primary/10 text-primary px-1.5 py-0.2 text-micro font-mono">
                  {historyData.entries.length}
                </span>
              ) : null}
            </TabsTrigger>
            <TabsTrigger value="json" data-testid="tab-json" className="flex items-center gap-2 text-xs sm:text-sm">
              <FileCode className="h-3.5 w-3.5" />
              Raw JSON
            </TabsTrigger>
          </TabsList>

          {/* Overview Tab Content */}
          <TabsContent value="overview" className="space-y-6 mt-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Main Fields Card (2 cols) */}
              <Card className="lg:col-span-2 border-border-subtle bg-surface-2">
                <CardHeader>
                  <CardTitle className="text-lead">Properties & Attributes</CardTitle>
                </CardHeader>
                <CardContent className="divide-y divide-border-subtle">
                  {metadata.fields
                    .filter((f) => f.name !== "id")
                    .map((field) => {
                      const value = objectData[field.name];
                      const isEmpty =
                        value === null || value === undefined || value === "";
                      const isBool =
                        field.type === "boolean" || typeof value === "boolean";
                      const isDate = isDateField(field.name, field.type);
                      const isChoice = Boolean(
                        field.choices && field.choices.length > 0
                      );
                      const matchedChoice = field.choices?.find(
                        (c) => String(c.value) === String(value)
                      );
                      const isComplex =
                        typeof value === "object" && value !== null;
                      const relation = relationByField.get(field.name);

                      return (
                        <div
                          key={field.name}
                          className="py-3.5 flex flex-col sm:flex-row sm:items-baseline justify-between gap-2"
                        >
                          <div className="sm:w-1/3 shrink-0">
                            <span className="text-micro font-semibold uppercase tracking-wider text-muted-foreground">
                              {field.label || field.name}
                            </span>
                            {field.help_text && (
                              <p className="text-meta text-muted-foreground/70 mt-0.5">
                                {field.help_text}
                              </p>
                            )}
                          </div>

                          <div className="sm:w-2/3 break-words text-body">
                            {isEmpty ? (
                              <EmptyValue />
                            ) : isBool ? (
                              value ? (
                                <StatusBadge tone="success">
                                  Yes
                                </StatusBadge>
                              ) : (
                                <StatusBadge tone="danger">
                                  No
                                </StatusBadge>
                              )
                            ) : isChoice ? (
                              <StatusBadge tone="neutral">
                                {matchedChoice?.label || String(value)}
                              </StatusBadge>
                            ) : isDate ? (
                              <span className="font-mono text-meta tabular-nums text-muted-foreground">
                                {formatDisplayDate(value)}
                              </span>
                            ) : relation ? (
                              <button
                                type="button"
                                onClick={() =>
                                  navigate({
                                    to: "/$model/$id/view",
                                    params: {
                                      model: relation.related_model,
                                      id: String(value),
                                    },
                                  })
                                }
                                className="font-mono text-meta tabular-nums text-muted-foreground transition-colors duration-fast ease-out hover:text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded"
                              >
                                #{String(value)}
                              </button>
                            ) : field.widget === "url" ||
                              (typeof value === "string" &&
                                value.startsWith("http")) ? (
                              <a
                                href={value}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1 text-primary hover:underline font-mono text-meta"
                              >
                                {value}
                                <ExternalLink className="h-3 w-3" />
                              </a>
                            ) : field.widget === "email" ? (
                              <a
                                href={`mailto:${value}`}
                                className="text-primary hover:underline font-mono text-meta"
                              >
                                {value}
                              </a>
                            ) : isComplex ? (
                              <pre className="rounded border border-border-subtle bg-surface-sunken p-3 text-meta font-mono whitespace-pre-wrap max-h-48 overflow-auto">
                                {JSON.stringify(value, null, 2)}
                              </pre>
                            ) : (
                              <span className="text-foreground font-medium">
                                {String(value)}
                              </span>
                            )}
                          </div>
                        </div>
                      );
                    })}
                </CardContent>
              </Card>

              {/* Sidebar Info Card (1 col) */}
              <div className="space-y-6">
                <Card className="border-border-subtle bg-surface-2">
                  <CardHeader>
                    <CardTitle className="text-lead">System Metadata</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4 text-meta">
                    <div className="flex items-center justify-between border-b border-border-subtle pb-2">
                      <span className="text-muted-foreground">Internal ID</span>
                      <span className="font-mono font-bold text-foreground tabular-nums">
                        {objectID}
                      </span>
                    </div>
                    <div className="flex items-center justify-between border-b border-border-subtle pb-2">
                      <span className="text-muted-foreground">Model Type</span>
                      <span className="font-mono text-foreground">
                        {metadata.name}
                      </span>
                    </div>
                    {objectData.created_at && (
                      <div className="flex items-center justify-between border-b border-border-subtle pb-2">
                        <span className="text-muted-foreground">Created</span>
                        <span className="font-mono text-foreground tabular-nums">
                          {formatDisplayDate(objectData.created_at)}
                        </span>
                      </div>
                    )}
                    {objectData.updated_at && (
                      <div className="flex items-center justify-between border-b border-border-subtle pb-2">
                        <span className="text-muted-foreground">Last Updated</span>
                        <span className="font-mono text-foreground tabular-nums">
                          {formatDisplayDate(objectData.updated_at)}
                        </span>
                      </div>
                    )}
                    <div className="flex items-center justify-between pt-1">
                      <span className="text-muted-foreground">Permissions</span>
                      <div className="flex gap-1">
                        {metadata.permissions.change && (
                           <Badge variant="outline" className="text-micro px-1.5 py-0">
                            Writable
                          </Badge>
                        )}
                        {metadata.permissions.delete && (
                          <Badge variant="outline" className="text-micro px-1.5 py-0 text-destructive">
                            Deletable
                          </Badge>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Relations Card if configured */}
                {metadata.relations && metadata.relations.length > 0 && (
                  <Card className="border-border-subtle bg-surface-2">
                    <CardHeader>
                      <CardTitle className="text-lead flex items-center gap-2">
                        <Link2 className="h-4 w-4 text-primary" />
                        Related Models
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {metadata.relations.map((relation) => (
                        <div
                          key={relation.name}
                          className="flex items-center justify-between p-2 rounded bg-surface-sunken hover:bg-muted/70 transition-colors"
                        >
                          <div>
                            <p className="text-meta font-semibold text-foreground">
                              {relation.label || relation.name}
                            </p>
                            <p className="text-micro text-muted-foreground uppercase">
                              {relation.type} &middot; {relation.related_model}
                            </p>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            asChild
                            className="h-7 px-2 text-meta text-primary"
                          >
                            <Link
                              to="/$model"
                              params={{ model: relation.related_model }}
                            >
                              Browse
                            </Link>
                          </Button>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                )}
              </div>
            </div>
          </TabsContent>

          {/* Audit History Tab Content */}
          <TabsContent value="history" className="mt-6">
            <Card className="border-border-subtle bg-surface-2">
              <CardContent className="p-6">
                <AuditHistoryViewer
                  entries={historyData?.entries}
                  isLoading={historyLoading}
                  modelName={metadata.verbose_name}
                  objectId={objectID}
                />
              </CardContent>
            </Card>
          </TabsContent>

          {/* Raw JSON Tab Content */}
          <TabsContent value="json" className="mt-6">
            <Card className="border-border-subtle bg-surface-2">
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-lead">Raw Entity Payload</CardTitle>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  data-testid="copy-json"
                  onClick={copyJSON}
                  className="gap-1.5 text-meta"
                >
                  {copiedJSON ? (
                    <>
                      <Check className="h-3.5 w-3.5 text-success" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-3.5 w-3.5" />
                      Copy JSON
                    </>
                  )}
                </Button>
              </CardHeader>
              <CardContent>
                <pre className="rounded border border-border-subtle bg-surface-sunken p-4 text-meta font-mono text-foreground/90 overflow-x-auto max-h-[600px]">
                  {JSON.stringify(objectData, null, 2)}
                </pre>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Delete Confirmation Dialog */}
        <ConfirmationDialog
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          testId="delete-record-dialog"
          title={`Delete ${metadata.verbose_name}`}
          description={`Are you sure you want to permanently delete record #${objectID}? This action cannot be undone.`}
          confirmLabel="Delete Record"
          variant="destructive"
          onConfirm={handleDelete}
        />
      </div>
    </AdminLayout>
  );
}
