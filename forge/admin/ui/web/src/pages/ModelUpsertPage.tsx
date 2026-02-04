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
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Checkbox } from "../components/ui/checkbox";
import { SearchableSelect } from "../components/ui/searchable-select";
import { Card, CardContent } from "../components/ui/card";
import { Loader2, Save, X, Plus } from "lucide-react";
import { useState, useEffect, useRef } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import { useUIComponent } from "../hooks/useUIComponent";
import { cn } from "../lib/utils";
import { useToast } from "../hooks/use-toast";

import { ConfirmationDialog } from "../components/ui/confirmation-dialog";

interface ModelFormPageProps {
  mode: "create" | "edit";
}

export default function ModelFormPage({ mode }: ModelFormPageProps) {
  const params = useParams({ strict: false }) as any;
  const { model, id } = params; // 'model' and 'id' are used to derive modelName and objectId
  const navigate = useNavigate();
  const { toast } = useToast();
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [fieldErrors, setFieldErrors] = useState<Record<string, string[]>>({});
  const [inlineData, setInlineData] = useState<Record<string, any[]>>({});
  const [showExitDialog, setShowExitDialog] = useState(false);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const inlineOriginalRef = useRef<Record<string, any[]>>({});

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

  useEffect(() => {
    if (objectData && mode === "edit") {
      setFormData(objectData);
    }
  }, [objectData, mode]);

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

  // Resolve Overrides
  const FormHeader = useUIComponent(
    metadata?.ui_overrides?.["form.header"] || "",
    "div"
  );
  const FormFooter = useUIComponent(
    metadata?.ui_overrides?.["form.footer"] || "",
    "div"
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});

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
    } catch (error: any) {
      console.error("Failed to save:", error);
      const errorData = error.response?.data;
      const errorMsg =
        errorData?.message || "Failed to save changes. Please try again.";

      if (errorData?.details) {
        setFieldErrors(errorData.details);
      }

      toast({
        title: "Error",
        description: errorMsg,
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

  const formatTimestamp = (timestamp?: string) => {
    if (!timestamp) return "Unknown time";
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) return timestamp;
    return date.toLocaleString();
  };

  const formatChangeStats = (changeStats?: string) => {
    if (!changeStats) return [];
    try {
      const parsed = JSON.parse(changeStats);
      if (Array.isArray(parsed)) {
        return parsed.map((item) => String(item));
      }
      if (parsed && typeof parsed === "object") {
        return Object.entries(parsed).map(([field, details]) => {
          if (details && typeof details === "object") {
            const detailObj = details as Record<string, any>;
            const fromValue =
              detailObj.from ?? detailObj.old ?? detailObj.before;
            const toValue = detailObj.to ?? detailObj.new ?? detailObj.after;
            if (fromValue !== undefined || toValue !== undefined) {
              return `${field}: ${fromValue ?? "∅"} → ${toValue ?? "∅"}`;
            }
          }
          return `${field}: ${String(details)}`;
        });
      }
      return [String(parsed)];
    } catch (error) {
      return [changeStats];
    }
  };

  const getActionLabel = (action?: string) => {
    switch (action) {
      case "add":
        return "Created";
      case "change":
        return "Updated";
      case "delete":
        return "Deleted";
      default:
        return action ? action : "Updated";
    }
  };

  const renderField = (field: any) => {
    if (field.read_only && mode === "create") return null;

    const value = formData[field.name] || "";
    const fieldOverride = metadata.ui_overrides?.[`field.${field.name}`];
    const CustomField = useUIComponent(fieldOverride || "", null as any);

    if (CustomField) {
      return (
        <CustomField
          field={field}
          value={value}
          onChange={(val: any) => handleChange(field.name, val)}
          metadata={metadata}
        />
      );
    }

    switch (field.type) {
      case "boolean":
        return (
          <div className="flex items-center space-x-3 p-3 rounded-lg border border-border/50 bg-muted/20 hover:bg-muted/30 transition-colors">
            <Checkbox
              id={field.name}
              checked={!!value}
              onChange={(e) => handleChange(field.name, e.target.checked)}
            />
            <label
              htmlFor={field.name}
              className="text-sm font-semibold leading-none cursor-pointer select-none"
            >
              {field.label}
            </label>
          </div>
        );

      case "text":
        if (field.widget === "textarea") {
          return (
            <textarea
              id={field.name}
              value={value}
              onChange={(e) => handleChange(field.name, e.target.value)}
              className="flex min-h-[120px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all"
              required={field.required}
            />
          );
        }
        return (
          <Input
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
          />
        );

      case "integer":
      case "float":
      case "decimal":
        return (
          <Input
            type="number"
            id={field.name}
            value={value}
            onChange={(e) =>
              handleChange(field.name, parseFloat(e.target.value))
            }
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
          />
        );

      case "date":
        return (
          <Input
            type="date"
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
          />
        );

      case "datetime":
        return (
          <Input
            type="datetime-local"
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
          />
        );

      case "choice":
        return (
          <select
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="flex h-10 w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all appearance-none"
            required={field.required}
          >
            <option value="">Select {field.label}</option>
            {field.choices?.map((choice: any) => (
              <option key={choice.value} value={choice.value}>
                {choice.label}
              </option>
            ))}
          </select>
        );

      case "password":
        return (
          <Input
            type="password"
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            autoComplete="new-password"
          />
        );

      case "json":
        return (
          <textarea
            id={field.name}
            value={
              typeof value === "object" ? JSON.stringify(value, null, 2) : value
            }
            onChange={(e) => {
              const val = e.target.value;
              handleChange(field.name, val);
              // Basic validation hint could go here
            }}
            className="flex min-h-[150px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all shadow-inner"
            placeholder="{}"
            required={field.required}
          />
        );

      case "foreign_key":
      case "one_to_one":
        return (
          <SearchableSelect
            model={field.related_model}
            value={value}
            onChange={(val) => handleChange(field.name, val)}
            placeholder={`Select ${field.label}...`}
            required={field.required}
          />
        );

      case "many_to_many":
        const m2mValue = Array.isArray(value) ? value : [];
        return (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2 mb-2">
              {m2mValue.map((item: any, idx: number) => (
                <div
                  key={item.id || idx}
                  className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary border border-primary/20 text-xs font-medium animate-in fade-in zoom-in-95"
                >
                  <span>
                    {item.name ||
                      item.title ||
                      item.label ||
                      `ID: ${item.id || item}`}
                  </span>
                  <button
                    type="button"
                    onClick={() => {
                      const newValue = m2mValue.filter(
                        (_: any, i: number) => i !== idx
                      );
                      handleChange(field.name, newValue);
                    }}
                    className="hover:text-destructive transition-colors"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
            <SearchableSelect
              model={field.related_model}
              value={null}
              onChange={(val) => {
                if (!val) return;
                // Avoid duplicates
                if (
                  m2mValue.some(
                    (v: any) => (typeof v === "object" ? v.id : v) === val
                  )
                )
                  return;
                handleChange(field.name, [...m2mValue, val]);
              }}
              placeholder={`Add ${field.label}...`}
            />
          </div>
        );

      default:
        return (
          <Input
            id={field.name}
            value={value}
            onChange={(e) => handleChange(field.name, e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
            required={field.required}
          />
        );
    }
  };

  const renderInlineField = (
    field: any,
    value: any,
    onChange: (val: any) => void
  ) => {
    const isReadOnly = field.read_only;
    switch (field.type) {
      case "boolean":
        return (
          <div className="flex items-center space-x-3 p-3 rounded-lg border border-border/50 bg-muted/20">
            <Checkbox
              id={field.name}
              checked={!!value}
              onChange={(e) => onChange(e.target.checked)}
              disabled={isReadOnly}
            />
            <label
              htmlFor={field.name}
              className="text-sm font-semibold leading-none cursor-pointer select-none"
            >
              {field.label}
            </label>
          </div>
        );
      case "text":
        if (field.widget === "textarea") {
          return (
            <textarea
              id={field.name}
              value={value || ""}
              onChange={(e) => onChange(e.target.value)}
              className="flex min-h-[120px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all"
              required={field.required}
              disabled={isReadOnly}
            />
          );
        }
        return (
          <Input
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      case "integer":
      case "float":
      case "decimal":
        return (
          <Input
            type="number"
            id={field.name}
            value={value ?? ""}
            onChange={(e) =>
              onChange(
                e.target.value === "" ? "" : parseFloat(e.target.value)
              )
            }
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      case "date":
        return (
          <Input
            type="date"
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      case "datetime":
        return (
          <Input
            type="datetime-local"
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      case "choice":
        return (
          <select
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="flex h-10 w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all appearance-none"
            required={field.required}
            disabled={isReadOnly}
          >
            <option value="">Select {field.label}</option>
            {field.choices?.map((choice: any) => (
              <option key={choice.value} value={choice.value}>
                {choice.label}
              </option>
            ))}
          </select>
        );
      case "password":
        return (
          <Input
            type="password"
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus-visible:ring-primary/20 focus-visible:border-primary transition-all"
            required={field.required}
            disabled={isReadOnly}
            autoComplete="new-password"
          />
        );
      case "json":
        return (
          <textarea
            id={field.name}
            value={
              typeof value === "object" ? JSON.stringify(value, null, 2) : value
            }
            onChange={(e) => onChange(e.target.value)}
            className="flex min-h-[150px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-xs font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-50 transition-all shadow-inner"
            placeholder="{}"
            required={field.required}
            disabled={isReadOnly}
          />
        );
      case "foreign_key":
      case "one_to_one":
        return (
          <SearchableSelect
            model={field.related_model}
            value={value}
            onChange={(val) => onChange(val)}
            placeholder={`Select ${field.label}...`}
            required={field.required}
          />
        );
      case "many_to_many": {
        const m2mValue = Array.isArray(value) ? value : [];
        return (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2 mb-2">
              {m2mValue.map((item: any, idx: number) => (
                <div
                  key={item.id || idx}
                  className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary border border-primary/20 text-xs font-medium animate-in fade-in zoom-in-95"
                >
                  <span>
                    {item.name ||
                      item.title ||
                      item.label ||
                      `ID: ${item.id || item}`}
                  </span>
                  <button
                    type="button"
                    onClick={() => {
                      const newValue = m2mValue.filter(
                        (_: any, i: number) => i !== idx
                      );
                      onChange(newValue);
                    }}
                    className="hover:text-destructive transition-colors"
                    disabled={isReadOnly}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
            <SearchableSelect
              model={field.related_model}
              value={null}
              onChange={(val) => {
                if (!val) return;
                if (
                  m2mValue.some(
                    (v: any) => (typeof v === "object" ? v.id : v) === val
                  )
                )
                  return;
                onChange([...m2mValue, val]);
              }}
              placeholder={`Add ${field.label}...`}
            />
          </div>
        );
      }
      default:
        return (
          <Input
            id={field.name}
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="rounded-lg border-border/50 bg-background/50 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
            required={field.required}
            disabled={isReadOnly}
          />
        );
    }
  };

  const getInlineFields = (
    inlineMeta: any,
    inlineConfig: any,
    relatedField?: string
  ) => {
    const allowedFields = inlineConfig?.fields?.length
      ? new Set(inlineConfig.fields)
      : null;
    return (inlineMeta?.fields ?? []).filter((field: any) => {
      if (field.name === "id") return false;
      if (relatedField && field.name === relatedField) return false;
      if (allowedFields && !allowedFields.has(field.name)) return false;
      if (field.read_only && mode === "create") return false;
      return true;
    });
  };

  return (
    <AdminLayout>
      <div className="space-y-6">
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
              variant="ghost"
              onClick={handleCancel}
              className="text-muted-foreground hover:text-foreground"
            >
              Cancel
            </Button>
            <Button
              data-testid="submit-button"
              onClick={handleSubmit}
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

        <Card className="glass-lite border-border/50 shadow-xl shadow-black/5 overflow-hidden max-w-4xl mx-auto">
          <CardContent className="p-8">
            <form onSubmit={handleSubmit} className="space-y-8">
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
                        <p className="text-[11px] text-muted-foreground px-1 leading-relaxed">
                          {field.help_text}
                        </p>
                      )}
                    </div>
                  );
                })}
              </div>

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
          <Card className="glass-lite border-border/50 shadow-xl shadow-black/5 overflow-hidden max-w-4xl mx-auto">
            <CardContent className="p-8 space-y-6">
              <div>
                <h2 className="text-xl font-semibold text-foreground/90">
                  History
                </h2>
                <p className="text-sm text-muted-foreground">
                  Track who changed this record and what was updated.
                </p>
              </div>
              {historyLoading ? (
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Loading history…</span>
                </div>
              ) : historyData?.entries?.length ? (
                <ul className="relative border-l border-border/50 pl-6 space-y-6">
                  {historyData.entries.map((entry) => {
                    const changes = formatChangeStats(entry.change_stats);
                    const userLabel =
                      entry.user_name ||
                      entry.user_id ||
                      "System";
                    return (
                      <li key={entry.id} className="relative">
                        <span className="absolute -left-[9px] top-1.5 h-4 w-4 rounded-full bg-primary/70 border border-background shadow" />
                        <div className="space-y-2">
                          <div className="flex flex-wrap items-center gap-2 text-sm">
                            <span className="font-semibold text-foreground">
                              {userLabel}
                            </span>
                            <span className="text-muted-foreground">
                              {getActionLabel(entry.action)}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              {formatTimestamp(entry.timestamp)}
                            </span>
                          </div>
                          {entry.object_repr && (
                            <p className="text-sm text-muted-foreground">
                              {entry.object_repr}
                            </p>
                          )}
                          {changes.length > 0 ? (
                            <ul className="space-y-1 text-sm text-foreground/80">
                              {changes.map((change, index) => (
                                <li key={`${entry.id}-change-${index}`}>
                                  {change}
                                </li>
                              ))}
                            </ul>
                          ) : (
                            <p className="text-sm text-muted-foreground">
                              No change details recorded.
                            </p>
                          )}
                        </div>
                      </li>
                    );
                  })}
                </ul>
              ) : (
                <p className="text-sm text-muted-foreground">
                  No history events recorded yet.
                </p>
              )}
            </CardContent>
          </Card>
        )}
        {inlineRelationDetails.map((detail, index) => {
          const inlineMeta = inlineMetadataQueries[index]?.data;
          const inlineRows = inlineData[detail.relation.name] ?? [];
          const inlineLabel =
            detail.inlineConfig.label ||
            detail.relation.label ||
            detail.relation.name;
          const allowMultiple = detail.inlineConfig.type === "one_to_many";
          const inlineFields = getInlineFields(
            inlineMeta,
            detail.inlineConfig,
            detail.relatedField
          );

          return (
            <Card
              key={detail.relation.name}
              className="glass-lite border-border/50 shadow-xl shadow-black/5 overflow-hidden max-w-4xl mx-auto"
            >
              <CardContent className="p-8 space-y-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-xl font-semibold text-foreground">
                      {inlineLabel}
                    </h2>
                    {detail.inlineConfig.type && (
                      <p className="text-xs text-muted-foreground uppercase tracking-widest mt-1">
                        {detail.inlineConfig.type.replace(/_/g, " ")}
                      </p>
                    )}
                  </div>
                  {allowMultiple && (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => addInlineRow(detail.relation.name)}
                      className="gap-2"
                    >
                      <Plus className="h-4 w-4" />
                      Add {inlineLabel}
                    </Button>
                  )}
                </div>

                {inlineFields.length === 0 && (
                  <p className="text-sm text-muted-foreground">
                    No editable fields configured for this inline relation.
                  </p>
                )}

                <div className="space-y-6">
                  {inlineRows.length === 0 && allowMultiple && (
                    <div className="rounded-lg border border-dashed border-border/60 p-6 text-center text-sm text-muted-foreground">
                      No {inlineLabel} added yet.
                    </div>
                  )}

                  {inlineRows.map((row, rowIndex) => (
                    <div
                      key={row?.id || rowIndex}
                      className="rounded-lg border border-border/50 bg-background/40 p-6 space-y-6"
                    >
                      <div className="flex items-center justify-between">
                        <h3 className="text-sm font-semibold text-foreground">
                          {inlineLabel} {rowIndex + 1}
                        </h3>
                        {allowMultiple && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() =>
                              removeInlineRow(detail.relation.name, rowIndex)
                            }
                            className="text-muted-foreground hover:text-destructive"
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        )}
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
                        {inlineFields.map((field: any) => (
                          <div key={field.name} className="space-y-2">
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
                            {renderInlineField(
                              field,
                              row?.[field.name],
                              (value) =>
                                updateInlineRow(
                                  detail.relation.name,
                                  rowIndex,
                                  field.name,
                                  value
                                )
                            )}
                            {field.help_text && (
                              <p className="text-[11px] text-muted-foreground px-1 leading-relaxed">
                                {field.help_text}
                              </p>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          );
        })}

        <ConfirmationDialog
          open={showExitDialog}
          onOpenChange={setShowExitDialog}
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
