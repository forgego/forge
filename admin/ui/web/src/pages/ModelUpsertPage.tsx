import { useParams, useNavigate } from "@tanstack/react-router";
import { useQueries } from "@tanstack/react-query";
import {
  useModelMetadata,
  useModelDetail,
  useCreateObject,
  useUpdateObject,
  useModelHistory,
  adminKeys,
} from "../api/hooks/adminHooks";
import { adminAPI } from "../api/client";
import { parseApiError } from "../api/errors";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import { Loader2, Save } from "lucide-react";
import { useState, useEffect, useMemo, useRef } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import { useUIComponent } from "../hooks/useUIComponent";
import { cn } from "../lib/utils";
import { useToast } from "../hooks/use-toast";

import { ConfirmationDialog } from "../components/ui/confirmation-dialog";
import { FieldRenderer } from "../components/form/FieldRenderer";
import { InlineRelations } from "../components/form/InlineRelations";
import { HistorySection } from "../components/form/HistorySection";

interface ModelFormPageProps {
  mode: "create" | "edit";
}

// Re-exported from their new home so existing import paths keep working.
export { resolveFieldKind, isIntegerField } from "../components/form/resolve-field-kind";
export type { FieldKind } from "../components/form/types";

export default function ModelFormPage({ mode }: ModelFormPageProps) {
  const params = useParams({ strict: false }) as any;
  const { model, id } = params; // 'model' and 'id' are used to derive modelName and objectId
  const navigate = useNavigate();
  const { toast } = useToast();
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, string[]>>({});
  const [formErrors, setFormErrors] = useState<string[]>([]);
  const [inlineData, setInlineData] = useState<Record<string, any[]>>({});
  const [showExitDialog, setShowExitDialog] = useState(false);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const inlineOriginalRef = useRef<Record<string, any[]>>({});
  // Submitting via requestSubmit() (instead of calling handleSubmit directly)
  // preserves native required/maxlength validation on the header save button.
  const formRef = useRef<HTMLFormElement>(null);

  const modelName = model as string;
  const objectId = (id as string) || "";

  const { data: metadata, isLoading: metaLoading } =
    useModelMetadata(modelName);

  const { data: objectData, isLoading: objectLoading } = useModelDetail(
    modelName,
    objectId,
    {
      enabled: mode === "edit" && !!objectId,
    } as any
  );
  const { data: historyData, isLoading: historyLoading } = useModelHistory(
    modelName,
    objectId,
    {
      enabled: mode === "edit" && !!objectId,
    } as any
  );

  const createMutation = useCreateObject(modelName);
  const updateMutation = useUpdateObject(modelName);

  const inlineRelations =
    metadata?.relations?.filter((relation) => relation.inline) ?? [];
  const inlineRelationDetails = inlineRelations.map((relation) => {
    const inlineConfig = relation.inline ?? {};
    return {
      relation,
      inlineConfig,
      relatedModel: inlineConfig.related_model || relation.related_model,
      relatedField: inlineConfig.related_field || relation.related_field,
    };
  });

  // Field name -> relation (resolves FK/M2M widgets for plain fields).
  const relationsByName = useMemo(
    () =>
      new Map(
        (metadata?.relations ?? []).map((relation: any) => [relation.name, relation])
      ),
    [metadata?.relations]
  );

  const displayedFormErrors = useMemo(() => {
    const errors = [...formErrors];
    const renderedFieldNames = new Set(
      metadata?.fields
        ?.filter((f) => !(f.name === "id" || (f.read_only && mode === "create")))
        ?.map((f) => f.name) ?? []
    );
    for (const [key, msgs] of Object.entries(fieldErrors)) {
      if (!renderedFieldNames.has(key)) {
        for (const msg of msgs) {
          errors.push(`${key}: ${msg}`);
        }
      }
    }
    return errors;
  }, [formErrors, fieldErrors, metadata?.fields, mode]);

  const inlineMetadataQueries = useQueries({
    queries: inlineRelationDetails.map((detail) => ({
      queryKey: adminKeys.modelMeta(detail.relatedModel),
      queryFn: () => adminAPI.getModelMetadata(detail.relatedModel),
      enabled: !!detail.relatedModel,
      staleTime: 5 * 60 * 1000,
    })),
  });

  const inlineListQueries = useQueries({
    queries: inlineRelationDetails.map((detail) => ({
      queryKey: adminKeys.modelList(
        detail.relatedModel,
        detail.relatedField && mode === "edit" && objectId
          ? { [detail.relatedField]: objectId, page_size: 250 }
          : undefined
      ),
      queryFn: () =>
        adminAPI.listObjects(detail.relatedModel, {
          [detail.relatedField || ""]: objectId,
          page_size: 250,
        }),
      enabled:
        mode === "edit" &&
        !!objectId &&
        !!detail.relatedModel &&
        !!detail.relatedField,
    })),
  });

  // Sync fetched object into the form when it (re)loads in edit mode.
  // Done during render (adjust-state pattern) instead of an effect.
  const [formSource, setFormSource] = useState<unknown>(null);
  if (mode === "edit" && objectData && formSource !== objectData) {
    setFormSource(objectData);
    setFormData(objectData);
  }

  useEffect(() => {
    inlineRelationDetails.forEach((detail, index) => {
      const relationKey = detail.relation.name;
      const listQuery = inlineListQueries[index];
      if (!listQuery?.data) return;
      if (inlineOriginalRef.current[relationKey]) return;
      const results = (listQuery.data as any)?.results ?? [];
      inlineOriginalRef.current[relationKey] = results;
      setInlineData((prev) => ({ ...prev, [relationKey]: results }));
    });
  }, [inlineListQueries, inlineRelationDetails]);

  const persistInlineRelations = async (parentId: string | number) => {
    if (!inlineRelationDetails.length) return;
    const operations: Promise<any>[] = [];

    inlineRelationDetails.forEach((detail) => {
      const relationKey = detail.relation.name;
      const relatedModel = detail.relatedModel;
      const relatedField = detail.relatedField;
      if (!relatedModel || !relatedField) return;

      const currentItems = inlineData[relationKey] ?? [];
      const originalItems = inlineOriginalRef.current[relationKey] ?? [];
      const originalIds = new Set(
        originalItems.map((item: any) => item?.id).filter(Boolean)
      );
      const currentIds = new Set(
        currentItems.map((item: any) => item?.id).filter(Boolean)
      );

      originalIds.forEach((id) => {
        if (!currentIds.has(id)) {
          operations.push(adminAPI.deleteObject(relatedModel, id));
        }
      });

      currentItems.forEach((item: any) => {
        const { id, ...rest } = item || {};
        const payload = { ...rest, [relatedField]: parentId };
        if (id) {
          operations.push(adminAPI.updateObject(relatedModel, id, payload));
        } else {
          operations.push(adminAPI.createObject(relatedModel, payload));
        }
      });
    });

    if (operations.length) {
      await Promise.all(operations);
    }
  };

  // Resolve Overrides (top-level: stable hook order)
  const FormHeader = useUIComponent(
    metadata?.ui_overrides?.["form.header"] || "",
    "div"
  );
  const FormFooter = useUIComponent(
    metadata?.ui_overrides?.["form.footer"] || "",
    "div"
  );
  const FormBody = useUIComponent(
    metadata?.ui_overrides?.["form.body"] || "",
    null as any
  );

  const validateJsonFields = (): boolean => {
    if (!metadata) return true;
    const errors: Record<string, string[]> = {};
    for (const field of metadata.fields) {
      if (field.type !== "json") continue;
      const raw = formData[field.name];
      if (raw === undefined || raw === null || raw === "") continue;
      if (typeof raw === "string") {
        try {
          JSON.parse(raw);
        } catch {
          errors[field.name] = ["Must be valid JSON"];
        }
      }
    }
    if (Object.keys(errors).length > 0) {
      setFieldErrors((prev) => ({ ...prev, ...errors }));
      toast({
        title: "Invalid JSON",
        description: "One or more JSON fields contain invalid JSON.",
        variant: "destructive",
      });
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setFormErrors([]);
    if (!validateJsonFields()) return;

    try {
      if (mode === "create") {
        const created = await createMutation.mutateAsync(formData);
        const createdId =
          created?.id ?? created?.ID ?? created?.Id ?? created?.pk;
        if (createdId) {
          await persistInlineRelations(createdId);
        }
        toast({
          title: "Success",
          description: `${metadata?.verbose_name} created successfully`,
        });
      } else {
        await updateMutation.mutateAsync({ id: objectId, data: formData });
        await persistInlineRelations(objectId);
        toast({
          title: "Success",
          description: `${metadata?.verbose_name} updated successfully`,
        });
      }
      setHasUnsavedChanges(false);
      navigate({ to: "/$model", params: { model: modelName } });
    } catch (error) {
      console.error("Failed to save:", error);
      const info = parseApiError(
        error,
        "Failed to save changes. Please try again."
      );
      setFieldErrors(info.fieldErrors);
      setFormErrors(info.formErrors);
      toast({
        title: "Could not save",
        description: info.message,
        variant: "destructive",
      });
    }
  };

  const handleChange = (name: string, value: any) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
    setHasUnsavedChanges(true);
    // Clear error for this field
    if (fieldErrors[name]) {
      setFieldErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const updateInlineRow = (
    relationName: string,
    rowIndex: number,
    fieldName: string,
    value: any
  ) => {
    setInlineData((prev) => {
      const rows = [...(prev[relationName] ?? [])];
      const row = { ...(rows[rowIndex] ?? {}) };
      row[fieldName] = value;
      rows[rowIndex] = row;
      return { ...prev, [relationName]: rows };
    });
    setHasUnsavedChanges(true);
  };

  const addInlineRow = (relationName: string) => {
    setInlineData((prev) => ({
      ...prev,
      [relationName]: [...(prev[relationName] ?? []), {}],
    }));
    setHasUnsavedChanges(true);
  };

  const removeInlineRow = (relationName: string, rowIndex: number) => {
    setInlineData((prev) => {
      const rows = [...(prev[relationName] ?? [])];
      rows.splice(rowIndex, 1);
      return { ...prev, [relationName]: rows };
    });
    setHasUnsavedChanges(true);
  };

  const handleCancel = () => {
    if (hasUnsavedChanges) {
      setShowExitDialog(true);
    } else {
      navigate({ to: "/$model", params: { model: modelName } });
    }
  };

  if (metaLoading || (mode === "edit" && objectLoading)) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (!metadata) return null;

  const renderField = (field: any) => (
    <FieldRenderer
      field={field}
      // Preserve 0/false: only null/undefined become "".
      value={formData[field.name] ?? ""}
      onChange={(val: any) => handleChange(field.name, val)}
      relation={relationsByName.get(field.name)}
      mode={mode}
      overrideKey={metadata.ui_overrides?.[`field.${field.name}`]}
      metadata={metadata}
    />
  );

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* eslint-disable-next-line react-hooks/static-components -- useUIComponent returns a stable registry ref */}
        <FormHeader className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground/90">
              {mode === "create" ? "Add" : "Edit"} {metadata.verbose_name}
            </h1>
            <p className="text-muted-foreground text-sm">
              {mode === "create"
                ? `Create a new instance of ${metadata.verbose_name}`
                : `Updating ${metadata.verbose_name} #${objectId}`}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              data-testid="cancel-button"
              variant="ghost"
              onClick={handleCancel}
              className="text-muted-foreground hover:text-foreground"
            >
              Cancel
            </Button>
            <Button
              type="button"
              data-testid="submit-button"
              onClick={() => formRef.current?.requestSubmit()}
              disabled={createMutation.isPending || updateMutation.isPending}
              className="bg-primary hover:bg-primary/90 min-w-[120px]"
            >
              {createMutation.isPending || updateMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <Save className="h-4 w-4 mr-2" />
              )}
              {mode === "create" ? "Create" : "Save Changes"}
            </Button>
          </div>
        </FormHeader>

        <Card className="overflow-hidden max-w-4xl mx-auto">
          <CardContent className="p-4 sm:p-8">
            <form ref={formRef} onSubmit={handleSubmit} className="space-y-8">
              {displayedFormErrors.length > 0 && (
                <div
                  role="alert"
                  data-testid="form-errors"
                  className="rounded-md border border-destructive/40 bg-destructive/10 px-4 py-3 text-ui text-destructive"
                >
                  <ul className="list-disc pl-5 space-y-1">
                    {displayedFormErrors.map((m) => (
                      <li key={m}>{m}</li>
                    ))}
                  </ul>
                </div>
              )}

              {FormBody ? (
                // eslint-disable-next-line react-hooks/static-components -- useUIComponent returns a stable registry ref
                <FormBody
                  fields={metadata.fields}
                  formData={formData}
                  errors={fieldErrors}
                  onChange={handleChange}
                  renderField={renderField}
                  metadata={metadata}
                  mode={mode}
                />
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
                  {metadata.fields.map((field) => {
                    if (field.read_only && mode === "create") return null;
                    if (field.name === "id") return null;

                    const isFullWidth =
                      field.type === "text" && field.widget === "textarea";

                    return (
                      <div
                        key={field.name}
                        className={cn(
                          "space-y-2",
                          isFullWidth && "md:col-span-2"
                        )}
                      >
                        <div className="flex items-center justify-between px-1">
                          <label
                            htmlFor={field.name}
                            className="text-xs font-bold uppercase tracking-widest text-muted-foreground/80"
                          >
                            {field.label}{" "}
                            {field.required && (
                              <span className="text-destructive font-normal">
                                *
                              </span>
                            )}
                          </label>
                        </div>
                        {renderField(field)}
                        {fieldErrors[field.name] && (
                          <p className="text-xs text-destructive font-medium animate-in slide-in-from-top-1">
                            {fieldErrors[field.name].join(", ")}
                          </p>
                        )}
                        {field.help_text && (
                          <p className="text-micro text-muted-foreground px-1 leading-relaxed">
                            {field.help_text}
                          </p>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}

              {/* eslint-disable-next-line react-hooks/static-components -- useUIComponent returns a stable registry ref */}
              <FormFooter className="pt-4 border-t border-border/50 flex justify-end">
                <Button
                  type="submit"
                  disabled={
                    createMutation.isPending || updateMutation.isPending
                  }
                  className="bg-primary hover:bg-primary/90 shadow-lg shadow-primary/20 min-w-[140px]"
                >
                  <Save className="h-4 w-4 mr-2" />
                  {mode === "create"
                    ? "Create " + metadata.verbose_name
                    : "Save Changes"}
                </Button>
              </FormFooter>
            </form>
          </CardContent>
        </Card>

        {mode === "edit" && (
          <HistorySection
            historyData={historyData}
            historyLoading={historyLoading}
          />
        )}
        <InlineRelations
          details={inlineRelationDetails}
          metaQueries={inlineMetadataQueries}
          inlineData={inlineData}
          mode={mode}
          onUpdateRow={updateInlineRow}
          onAddRow={addInlineRow}
          onRemoveRow={removeInlineRow}
        />

        <ConfirmationDialog
          open={showExitDialog}
          onOpenChange={setShowExitDialog}
          testId="exit-dialog"
          title="Unsaved Changes"
          description="You have unsaved changes. Are you sure you want to leave? Your changes will be lost."
          confirmLabel="Leave"
          variant="destructive"
          onConfirm={() =>
            navigate({ to: "/$model", params: { model: modelName } })
          }
        />
      </div>
    </AdminLayout>
  );
}
