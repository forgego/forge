import { useState } from "react";
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
  CheckCircle2,
  XCircle,
  Link2,
} from "lucide-react";
import AdminLayout from "../components/layout/AdminLayout";
import {
  Card,
  CardContent,
  CardDescription,
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
import { cn } from "../lib/utils";

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
          <p className="text-sm text-muted-foreground animate-pulse">
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
            <h2 className="text-2xl font-bold tracking-tight text-foreground">
              Unable to load record
            </h2>
            <p className="text-sm text-muted-foreground mt-1 max-w-md">
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
        {/* Top Header & Actions */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border/40 pb-6">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-xs text-muted-foreground hover:text-foreground -ml-2"
                onClick={() => navigate({ to: "/$model", params: { model: modelName } })}
              >
                <ArrowLeft className="h-3.5 w-3.5 mr-1" />
                {metadata.verbose_name_plural}
              </Button>
              <span className="text-muted-foreground/40">/</span>
              <Badge variant="outline" className="text-xs font-mono">
                {metadata.name} #{objectID}
              </Badge>
              {statusValue !== undefined && (
                <Badge
                  variant={
                    statusValue === true ||
                    statusValue === "active" ||
                    statusValue === "published" ||
                    statusValue === "completed" ||
                    statusValue === "enabled" ||
                    statusValue === "open" ||
                    statusValue === "paid" ||
                    statusValue === "approved"
                      ? "default"
                      : "secondary"
                  }
                  className="capitalize text-xs font-medium"
                >
                  {String(statusValue)}
                </Badge>
              )}
            </div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              {primaryLabel}
            </h1>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            {metadata.permissions.change && (
              <Button
                data-testid="edit-record"
                onClick={() =>
                  navigate({
                    to: "/$model/$id",
                    params: { model: modelName, id: objectID },
                  })
                }
                className="shadow-sm"
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
          </div>
        </div>

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
                <span className="ml-1 rounded-full bg-primary/10 text-primary px-1.5 py-0.2 text-[10px] font-mono">
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
              <Card className="lg:col-span-2 border-border/50 shadow-sm">
                <CardHeader>
                  <CardTitle className="text-lg">Properties & Attributes</CardTitle>
                  <CardDescription>
                    All registered schema properties for this {metadata.verbose_name.toLowerCase()}.
                  </CardDescription>
                </CardHeader>
                <CardContent className="divide-y divide-border/40">
                  {metadata.fields
                    .filter((f) => f.name !== "id")
                    .map((field) => {
                      const value = objectData[field.name];
                      const hasValue =
                        value !== null && value !== undefined && value !== "";
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

                      return (
                        <div
                          key={field.name}
                          className="py-3.5 flex flex-col sm:flex-row sm:items-baseline justify-between gap-2"
                        >
                          <div className="sm:w-1/3 shrink-0">
                            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                              {field.label || field.name}
                            </span>
                            {field.help_text && (
                              <p className="text-[11px] text-muted-foreground/70 mt-0.5">
                                {field.help_text}
                              </p>
                            )}
                          </div>

                          <div className="sm:w-2/3 break-words text-sm">
                            {!hasValue ? (
                              <span className="text-muted-foreground/40 italic font-mono text-xs">
                                — (null)
                              </span>
                            ) : isBool ? (
                              <Badge
                                variant={value ? "default" : "outline"}
                                className={cn(
                                  "text-xs inline-flex items-center gap-1",
                                  value
                                    ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                                    : "text-muted-foreground"
                                )}
                              >
                                {value ? (
                                  <>
                                    <CheckCircle2 className="h-3 w-3" /> Yes
                                  </>
                                ) : (
                                  <>
                                    <XCircle className="h-3 w-3" /> No
                                  </>
                                )}
                              </Badge>
                            ) : isChoice ? (
                              <Badge variant="secondary" className="font-medium text-xs">
                                {matchedChoice?.label || String(value)}
                              </Badge>
                            ) : isDate ? (
                              <span className="font-mono text-xs text-foreground/90">
                                {formatDisplayDate(value)}
                              </span>
                            ) : field.widget === "url" ||
                              (typeof value === "string" &&
                                value.startsWith("http")) ? (
                              <a
                                href={value}
                                target="_blank"
                                rel="noreferrer"
                                className="inline-flex items-center gap-1 text-primary hover:underline font-mono text-xs"
                              >
                                {value}
                                <ExternalLink className="h-3 w-3" />
                              </a>
                            ) : field.widget === "email" ? (
                              <a
                                href={`mailto:${value}`}
                                className="text-primary hover:underline font-mono text-xs"
                              >
                                {value}
                              </a>
                            ) : isComplex ? (
                              <pre className="rounded-lg border border-border/50 bg-muted/40 p-3 text-xs font-mono whitespace-pre-wrap max-h-48 overflow-auto">
                                {JSON.stringify(value, null, 2)}
                              </pre>
                            ) : (
                              <span className="text-foreground/90 font-medium">
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
                <Card className="border-border/50 shadow-sm">
                  <CardHeader>
                    <CardTitle className="text-base">System Metadata</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4 text-xs">
                    <div className="flex items-center justify-between border-b border-border/40 pb-2">
                      <span className="text-muted-foreground">Internal ID</span>
                      <span className="font-mono font-bold text-foreground">
                        {objectID}
                      </span>
                    </div>
                    <div className="flex items-center justify-between border-b border-border/40 pb-2">
                      <span className="text-muted-foreground">Model Type</span>
                      <span className="font-mono text-foreground">
                        {metadata.name}
                      </span>
                    </div>
                    {objectData.created_at && (
                      <div className="flex items-center justify-between border-b border-border/40 pb-2">
                        <span className="text-muted-foreground">Created</span>
                        <span className="font-mono text-foreground">
                          {formatDisplayDate(objectData.created_at)}
                        </span>
                      </div>
                    )}
                    {objectData.updated_at && (
                      <div className="flex items-center justify-between border-b border-border/40 pb-2">
                        <span className="text-muted-foreground">Last Updated</span>
                        <span className="font-mono text-foreground">
                          {formatDisplayDate(objectData.updated_at)}
                        </span>
                      </div>
                    )}
                    <div className="flex items-center justify-between pt-1">
                      <span className="text-muted-foreground">Permissions</span>
                      <div className="flex gap-1">
                        {metadata.permissions.change && (
                          <Badge variant="outline" className="text-[10px] px-1.5 py-0">
                            Writable
                          </Badge>
                        )}
                        {metadata.permissions.delete && (
                          <Badge variant="outline" className="text-[10px] px-1.5 py-0 text-destructive">
                            Deletable
                          </Badge>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Relations Card if configured */}
                {metadata.relations && metadata.relations.length > 0 && (
                  <Card className="border-border/50 shadow-sm">
                    <CardHeader>
                      <CardTitle className="text-base flex items-center gap-2">
                        <Link2 className="h-4 w-4 text-primary" />
                        Related Models
                      </CardTitle>
                      <CardDescription>
                        Direct links to linked relations in the schema.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {metadata.relations.map((relation) => (
                        <div
                          key={relation.name}
                          className="flex items-center justify-between p-2 rounded-lg bg-muted/40 hover:bg-muted/70 transition-colors"
                        >
                          <div>
                            <p className="text-xs font-semibold text-foreground">
                              {relation.label || relation.name}
                            </p>
                            <p className="text-[10px] text-muted-foreground uppercase">
                              {relation.type} &middot; {relation.related_model}
                            </p>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            asChild
                            className="h-7 px-2 text-xs text-primary"
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
            <Card className="border-border/50 shadow-sm">
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
            <Card className="border-border/50 shadow-sm">
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-lg">Raw Entity Payload</CardTitle>
                  <CardDescription>
                    Direct JSON object representation received from the Admin API.
                  </CardDescription>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  data-testid="copy-json"
                  onClick={copyJSON}
                  className="gap-1.5 text-xs"
                >
                  {copiedJSON ? (
                    <>
                      <Check className="h-3.5 w-3.5 text-emerald-500" />
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
                <pre className="rounded-xl border border-border/60 bg-muted/30 p-4 text-xs font-mono text-foreground/90 overflow-x-auto max-h-[600px]">
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
